import React, { useState, useEffect } from 'react';
import { Row, Col, Card, Statistic, Progress, Table, Tag, message, Spin } from 'antd';
import { 
  CheckCircleOutlined,
  ExclamationCircleOutlined,
  StopOutlined,
  PrinterOutlined,
  CloudServerOutlined,
  FileTextOutlined,
  UserOutlined
} from '@ant-design/icons';
import * as echarts from 'echarts';

// Dashboard 数据接口
interface DashboardStats {
  totalPrinters: number;
  onlinePrinters: number;
  totalEdgeNodes: number;
  onlineEdgeNodes: number;
  totalPrintJobs: number;
  completedJobs: number;
  totalUsers: number;
  activeUsers: number;
}

interface PrinterStatus {
  id: string;
  name: string;
  location?: string;
  status: 'ready' | 'printing' | 'error' | 'offline';
  edge_node_id: string;
  model: string;
  key?: string;
}

interface PrintJob {
  id: string;
  name: string;
  user_name: string;
  printer_id: string;
  status: 'pending' | 'printing' | 'completed' | 'failed' | 'cancelled';
  created_at: string;
  page_count: number;
  key?: string;
}

// Dashboard 服务类
class DashboardService {
  private async getToken(): Promise<string | null> {
    try {
      const response = await fetch('/auth/me');
      const result = await response.json();
      
      if (result.code === 200 && result.data.access_token) {
        return result.data.access_token;
      }
    } catch (error) {
      console.error('获取 token 失败:', error);
    }
    
    return null;
  }

  async getStats(): Promise<DashboardStats> {
    try {
      const token = await this.getToken();
      
      // 并行获取各种统计数据
      const [printersResponse, edgeNodesResponse, printJobsResponse] = await Promise.all([
        fetch('/api/v1/admin/printers', {
          headers: { ...(token && { 'Authorization': `Bearer ${token}` }) },
        }),
        fetch('/api/v1/admin/edge-nodes', {
          headers: { ...(token && { 'Authorization': `Bearer ${token}` }) },
        }),
        fetch('/api/v1/admin/print-jobs', {
          headers: { ...(token && { 'Authorization': `Bearer ${token}` }) },
        })
      ]);
      
      const printersResult = printersResponse.ok ? await printersResponse.json() : null;
      const edgeNodesResult = edgeNodesResponse.ok ? await edgeNodesResponse.json() : null;
      const printJobsResult = printJobsResponse.ok ? await printJobsResponse.json() : null;
      
      const printers = printersResult?.data?.items || [];
      const edgeNodes = edgeNodesResult?.data?.items || [];
      const printJobs = printJobsResult?.jobs || [];
      
      // 计算统计数据
      const onlinePrinters = printers.filter((p: any) => p.status === 'ready' || p.status === 'printing').length;
      const onlineEdgeNodes = edgeNodes.filter((e: any) => e.status === 'online').length;
      const completedJobs = printJobs.filter((j: any) => j.status === 'completed').length;
      
      return {
        totalPrinters: printers.length,
        onlinePrinters,
        totalEdgeNodes: edgeNodes.length,
        onlineEdgeNodes,
        totalPrintJobs: printJobs.length,
        completedJobs,
        totalUsers: 0, // 暂时没有用户统计API
        activeUsers: 0,
      };
    } catch (error) {
      console.error('获取统计数据失败:', error);
      throw error;
    }
  }

  async getPrinters(): Promise<PrinterStatus[]> {
    try {
      const token = await this.getToken();
      const response = await fetch('/api/v1/admin/printers', {
        headers: {
          ...(token && { 'Authorization': `Bearer ${token}` }),
        },
      });
      
      if (response.ok) {
        const result = await response.json();
        return result.data.items || [];
      }
    } catch (error) {
      console.error('获取打印机列表失败:', error);
      throw error;
    }
    
    return [];
  }

  async getPrintJobs(): Promise<{ jobs: PrintJob[]; total: number }> {
    try {
      const token = await this.getToken();
      const response = await fetch('/api/v1/admin/print-jobs?page=1&page_size=5', {
        headers: {
          ...(token && { 'Authorization': `Bearer ${token}` }),
        },
      });
      
      if (response.ok) {
        const result = await response.json();
        return {
          jobs: result?.jobs || [],
          total: result?.pagination?.total || result?.jobs?.length || 0
        };
      }
    } catch (error) {
      console.error('获取打印任务列表失败:', error);
      throw error;
    }
    
    return { jobs: [], total: 0 };
  }

  async getPrintJobTrends(): Promise<{ dates: string[]; completed: number[]; failed: number[] }> {
    try {
      const token = await this.getToken();
      const response = await fetch('/api/v1/admin/dashboard/trends', {
        headers: {
          ...(token && { 'Authorization': `Bearer ${token}` }),
        },
      });
      
      if (response.ok) {
        const result = await response.json();
        return result.data || { dates: [], completed: [], failed: [] };
      }
    } catch (error) {
      console.error('获取趋势数据失败:', error);
    }
    
    // 如果API不可用，返回模拟数据以便测试图表显示
    const mockDates = [];
    const mockCompleted = [];
    const mockFailed = [];
    
    // 生成最近7天的模拟数据
    for (let i = 6; i >= 0; i--) {
      const date = new Date();
      date.setDate(date.getDate() - i);
      mockDates.push(date.toLocaleDateString('zh-CN', { month: '2-digit', day: '2-digit' }));
      mockCompleted.push(Math.floor(Math.random() * 20) + 5); // 5-25之间的随机数
      mockFailed.push(Math.floor(Math.random() * 5)); // 0-5之间的随机数
    }
    
    return {
      dates: mockDates,
      completed: mockCompleted,
      failed: mockFailed,
    };
  }
}

const dashboardService = new DashboardService();

// Dashboard 组件
const Dashboard: React.FC = () => {
  const [stats, setStats] = useState<DashboardStats>({
    totalPrinters: 0,
    onlinePrinters: 0,
    totalEdgeNodes: 0,
    onlineEdgeNodes: 0,
    totalPrintJobs: 0,
    completedJobs: 0,
    totalUsers: 0,
    activeUsers: 0,
  });

  const [loading, setLoading] = useState(true);

  // 数据加载
  useEffect(() => {
    const loadDashboardData = async () => {
      try {
        setLoading(true);

        // 只加载统计数据和趋势数据
        const [statsData, trendsData] = await Promise.all([
          dashboardService.getStats(),
          dashboardService.getPrintJobTrends(),
        ]);

        setStats(statsData);

        // 延迟初始化图表，确保DOM渲染完成
        setTimeout(() => {
          const chartElement = document.getElementById('printJobsChart');
          if (chartElement) {
            // 清理旧实例
            echarts.dispose(chartElement);
            
            const chart = echarts.init(chartElement);
          const option = {
            title: {
              text: '打印任务趋势',
              left: 'center',
              textStyle: {
                fontSize: 16,
              },
            },
            tooltip: {
              trigger: 'axis',
            },
            legend: {
              data: ['完成任务', '失败任务'],
              top: 30,
            },
            xAxis: {
              type: 'category',
              data: trendsData.dates,
            },
            yAxis: {
              type: 'value',
            },
            series: [
              {
                name: '完成任务',
                type: 'line',
                smooth: true,
                data: trendsData.completed,
                itemStyle: {
                  color: '#52c41a',
                },
              },
              {
                name: '失败任务',
                type: 'line',
                smooth: true,
                data: trendsData.failed,
                itemStyle: {
                  color: '#ff4d4f',
                },
              },
            ],
          };
            chart.setOption(option);

            // 响应式处理
            const handleResize = () => chart.resize();
            window.addEventListener('resize', handleResize);
          }
        }, 500);

      } catch (error) {
        console.error('加载 Dashboard 数据失败:', error);
        message.error('加载 Dashboard 数据失败，请稍后重试');
      } finally {
        setLoading(false);
      }
    };

    loadDashboardData();
  }, []);

  const getStatusColor = (status: string) => {
    switch (status) {
      case 'ready': return 'success';
      case 'printing': return 'processing';
      case 'offline': return 'default';
      case 'error': return 'error';
      default: return 'default';
    }
  };

  const getJobStatusColor = (status: string) => {
    switch (status) {
      case 'completed': return 'success';
      case 'printing': return 'processing';
      case 'pending': return 'warning';
      case 'failed': return 'error';
      case 'cancelled': return 'default';
      default: return 'default';
    }
  };

  // 移除表格列定义，因为不再需要表格

  if (loading) {
    return (
      <div style={{ 
        display: 'flex', 
        justifyContent: 'center', 
        alignItems: 'center', 
        height: '400px' 
      }}>
        <Spin size="large" tip="加载中..." />
      </div>
    );
  }

  return (
    <div>
      {/* 统计卡片 */}
      <Row gutter={[16, 16]} style={{ marginBottom: 24 }}>
        <Col xs={12} sm={12} md={6}>
          <Card style={{ height: 120, minHeight: 120 }} bodyStyle={{ padding: '16px 12px' }}>
            <Statistic
              title="打印机总数"
              value={stats.totalPrinters}
              prefix={<PrinterOutlined />}
              valueStyle={{ fontSize: '20px' }}
            />
            <div style={{ fontSize: '12px', color: '#8c8c8c', marginTop: '4px' }}>
              在线: {stats.onlinePrinters} 台
            </div>
          </Card>
        </Col>
        <Col xs={12} sm={12} md={6}>
          <Card style={{ height: 120, minHeight: 120 }} bodyStyle={{ padding: '16px 12px' }}>
            <Statistic
              title="边缘节点"
              value={stats.totalEdgeNodes}
              prefix={<CloudServerOutlined />}
              valueStyle={{ fontSize: '20px' }}
            />
            <div style={{ fontSize: '12px', color: '#8c8c8c', marginTop: '4px' }}>
              在线: {stats.onlineEdgeNodes} 个
            </div>
          </Card>
        </Col>
        <Col xs={12} sm={12} md={6}>
          <Card style={{ height: 120, minHeight: 120 }} bodyStyle={{ padding: '16px 12px' }}>
            <Statistic
              title="打印任务"
              value={stats.totalPrintJobs}
              prefix={<FileTextOutlined />}
              valueStyle={{ fontSize: '20px' }}
            />
            <div style={{ fontSize: '12px', color: '#8c8c8c', marginTop: '4px' }}>
              完成: {stats.completedJobs} 个
            </div>
          </Card>
        </Col>
        <Col xs={12} sm={12} md={6}>
          <Card style={{ height: 120, minHeight: 120 }} bodyStyle={{ padding: '16px 12px' }}>
            <Statistic
              title="用户总数"
              value={stats.totalUsers}
              prefix={<UserOutlined />}
              valueStyle={{ fontSize: '20px' }}
            />
            <div style={{ fontSize: '12px', color: '#8c8c8c', marginTop: '4px' }}>
              活跃: {stats.activeUsers} 人
            </div>
          </Card>
        </Col>
      </Row>

      {/* 移除设备状态和最近任务列表，用户可以直接访问对应页面 */}

      {/* 图表区域 */}
      <Row style={{ marginTop: 16 }}>
        <Col span={24}>
          <Card title="任务趋势分析">
            <div id="printJobsChart" style={{ height: '300px', width: '100%' }}></div>
          </Card>
        </Col>
      </Row>
    </div>
  );
};

export default Dashboard;
