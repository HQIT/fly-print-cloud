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
    // 暂时返回空数据，趋势图功能待后续开发
    return {
      dates: [],
      completed: [],
      failed: [],
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

  const [printers, setPrinters] = useState<PrinterStatus[]>([]);
  const [printJobs, setPrintJobs] = useState<PrintJob[]>([]);
  const [loading, setLoading] = useState(true);

  // 数据加载
  useEffect(() => {
    const loadDashboardData = async () => {
      try {
        setLoading(true);

        // 并行加载所有数据
        const [statsData, printersData, jobsData, trendsData] = await Promise.all([
          dashboardService.getStats(),
          dashboardService.getPrinters(),
          dashboardService.getPrintJobs(),
          dashboardService.getPrintJobTrends(),
        ]);

        setStats(statsData);
        setPrinters(printersData.map(p => ({ ...p, key: p.id })));
        setPrintJobs(jobsData.jobs.map(j => ({ ...j, key: j.id })));

        // 初始化图表
        const chartElement = document.getElementById('printJobsChart');
        if (chartElement) {
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
          
          return () => {
            window.removeEventListener('resize', handleResize);
            chart.dispose();
          };
        }

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

  const printerColumns = [
    {
      title: '打印机',
      dataIndex: 'name',
      key: 'name',
      ellipsis: true,
      width: 150,
    },
    {
      title: '位置',
      dataIndex: 'location',
      key: 'location',
      ellipsis: true,
      width: 120,
    },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      width: 80,
      render: (status: string) => (
        <Tag color={getStatusColor(status)} style={{ fontSize: '11px', padding: '0 4px' }}>
          {status === 'ready' ? '就绪' :
           status === 'printing' ? '打印中' :
           status === 'offline' ? '离线' : '错误'}
        </Tag>
      ),
    },
  ];

  const jobColumns = [
    {
      title: '文件名',
      dataIndex: 'name',
      key: 'name',
      ellipsis: true,
      width: 120,
    },
    {
      title: '用户',
      dataIndex: 'user_name',
      key: 'user_name',
      width: 60,
      ellipsis: true,
    },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      width: 70,
      render: (status: string) => (
        <Tag color={getJobStatusColor(status)} style={{ fontSize: '11px', padding: '0 4px' }}>
          {status === 'completed' ? '完成' :
           status === 'printing' ? '打印中' :
           status === 'pending' ? '等待' : 
           status === 'failed' ? '失败' :
           status === 'cancelled' ? '已取消' : status}
        </Tag>
      ),
    },
    {
      title: '页数',
      dataIndex: 'page_count',
      key: 'page_count',
      width: 50,
    },
  ];

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

      {/* 内容区域 */}
      <Row gutter={[16, 16]}>
        {/* 设备状态 */}
        <Col xs={24} lg={12}>
          <Card title="设备状态" style={{ height: '400px' }}>
            <Table
              columns={printerColumns}
              dataSource={printers}
              pagination={false}
              size="small"
              scroll={{ y: 280, x: 400 }}
            />
          </Card>
        </Col>

        {/* 任务状态 */}
        <Col xs={24} lg={12}>
          <Card title="最近任务" style={{ height: '400px' }}>
            <Table
              columns={jobColumns}
              dataSource={printJobs}
              pagination={false}
              size="small"
              scroll={{ y: 280, x: 300 }}
            />
          </Card>
        </Col>
      </Row>

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
