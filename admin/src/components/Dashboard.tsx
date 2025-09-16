import React, { useState, useEffect } from 'react';
import { Row, Col, Card, Statistic, Progress, Table, Tag, Space, Alert } from 'antd';
import { 
  PrinterOutlined, 
  CloudServerOutlined, 
  FileTextOutlined, 
  UserOutlined,
  CheckCircleOutlined,
  ExclamationCircleOutlined,
  ClockCircleOutlined,
  StopOutlined
} from '@ant-design/icons';
import * as echarts from 'echarts';

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
  key: string;
  name: string;
  location: string;
  status: 'online' | 'offline' | 'printing' | 'error';
  paperLevel: number;
  inkLevel: number;
}

interface PrintJob {
  key: string;
  id: string;
  fileName: string;
  user: string;
  printer: string;
  status: 'pending' | 'printing' | 'completed' | 'failed';
  createdAt: string;
  pages: number;
}

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

  const [printers] = useState<PrinterStatus[]>([
    {
      key: '1',
      name: 'HP LaserJet Pro M404n',
      location: '办公室 A-101',
      status: 'online',
      paperLevel: 85,
      inkLevel: 60,
    },
    {
      key: '2',
      name: 'Canon PIXMA G3020',
      location: '办公室 B-205',
      status: 'printing',
      paperLevel: 45,
      inkLevel: 30,
    },
    {
      key: '3',
      name: 'Epson EcoTank L3150',
      location: '会议室 C-301',
      status: 'offline',
      paperLevel: 20,
      inkLevel: 80,
    },
    {
      key: '4',
      name: 'Brother HL-L2350DW',
      location: '前台接待',
      status: 'error',
      paperLevel: 0,
      inkLevel: 15,
    },
  ]);

  const [printJobs] = useState<PrintJob[]>([
    {
      key: '1',
      id: 'JOB-2024-001',
      fileName: '合同文件.pdf',
      user: '张三',
      printer: 'HP LaserJet Pro M404n',
      status: 'completed',
      createdAt: '2024-01-15 10:30',
      pages: 5,
    },
    {
      key: '2',
      id: 'JOB-2024-002',
      fileName: '报告.docx',
      user: '李四',
      printer: 'Canon PIXMA G3020',
      status: 'printing',
      createdAt: '2024-01-15 11:15',
      pages: 12,
    },
    {
      key: '3',
      id: 'JOB-2024-003',
      fileName: '图表分析.xlsx',
      user: '王五',
      printer: 'Epson EcoTank L3150',
      status: 'pending',
      createdAt: '2024-01-15 11:45',
      pages: 3,
    },
    {
      key: '4',
      id: 'JOB-2024-004',
      fileName: '项目计划.pptx',
      user: '赵六',
      printer: 'Brother HL-L2350DW',
      status: 'failed',
      createdAt: '2024-01-15 12:00',
      pages: 20,
    },
  ]);

  // 模拟数据加载
  useEffect(() => {
    // 模拟 API 调用
    setTimeout(() => {
      setStats({
        totalPrinters: 12,
        onlinePrinters: 9,
        totalEdgeNodes: 5,
        onlineEdgeNodes: 4,
        totalPrintJobs: 156,
        completedJobs: 142,
        totalUsers: 48,
        activeUsers: 23,
      });
    }, 1000);

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
          data: ['周一', '周二', '周三', '周四', '周五', '周六', '周日'],
        },
        yAxis: {
          type: 'value',
        },
        series: [
          {
            name: '完成任务',
            type: 'line',
            smooth: true,
            data: [12, 18, 15, 22, 28, 16, 20],
            itemStyle: {
              color: '#52c41a',
            },
          },
          {
            name: '失败任务',
            type: 'line',
            smooth: true,
            data: [2, 1, 3, 2, 1, 4, 2],
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
  }, []);

  const getStatusColor = (status: string) => {
    switch (status) {
      case 'online': return 'success';
      case 'printing': return 'processing';
      case 'offline': return 'default';
      case 'error': return 'error';
      case 'completed': return 'success';
      case 'pending': return 'warning';
      case 'failed': return 'error';
      default: return 'default';
    }
  };

  const getStatusText = (status: string) => {
    switch (status) {
      case 'online': return '在线';
      case 'printing': return '打印中';
      case 'offline': return '离线';
      case 'error': return '故障';
      case 'completed': return '已完成';
      case 'pending': return '等待中';
      case 'failed': return '失败';
      default: return status;
    }
  };

  const printerColumns = [
    {
      title: '打印机名称',
      dataIndex: 'name',
      key: 'name',
    },
    {
      title: '位置',
      dataIndex: 'location',
      key: 'location',
    },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      render: (status: string) => (
        <Tag color={getStatusColor(status)}>{getStatusText(status)}</Tag>
      ),
    },
    {
      title: '纸张余量',
      dataIndex: 'paperLevel',
      key: 'paperLevel',
      render: (level: number) => (
        <Progress 
          percent={level} 
          size="small" 
          status={level < 20 ? 'exception' : level < 50 ? 'active' : 'success'}
        />
      ),
    },
    {
      title: '墨量余量',
      dataIndex: 'inkLevel',
      key: 'inkLevel',
      render: (level: number) => (
        <Progress 
          percent={level} 
          size="small" 
          status={level < 20 ? 'exception' : level < 50 ? 'active' : 'success'}
        />
      ),
    },
  ];

  const jobColumns = [
    {
      title: '任务ID',
      dataIndex: 'id',
      key: 'id',
    },
    {
      title: '文件名',
      dataIndex: 'fileName',
      key: 'fileName',
    },
    {
      title: '用户',
      dataIndex: 'user',
      key: 'user',
    },
    {
      title: '打印机',
      dataIndex: 'printer',
      key: 'printer',
    },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      render: (status: string) => (
        <Tag color={getStatusColor(status)}>{getStatusText(status)}</Tag>
      ),
    },
    {
      title: '页数',
      dataIndex: 'pages',
      key: 'pages',
    },
    {
      title: '创建时间',
      dataIndex: 'createdAt',
      key: 'createdAt',
    },
  ];

  return (
    <div>
      {/* 统计卡片 */}
      <Row gutter={16} style={{ marginBottom: 24 }}>
        <Col xs={24} sm={12} md={6}>
          <Card>
            <Statistic
              title="打印机总数"
              value={stats.totalPrinters}
              prefix={<PrinterOutlined />}
              suffix={`/ ${stats.onlinePrinters} 在线`}
              valueStyle={{ color: '#1890ff' }}
            />
          </Card>
        </Col>
        <Col xs={24} sm={12} md={6}>
          <Card>
            <Statistic
              title="边缘节点"
              value={stats.totalEdgeNodes}
              prefix={<CloudServerOutlined />}
              suffix={`/ ${stats.onlineEdgeNodes} 在线`}
              valueStyle={{ color: '#52c41a' }}
            />
          </Card>
        </Col>
        <Col xs={24} sm={12} md={6}>
          <Card>
            <Statistic
              title="打印任务"
              value={stats.totalPrintJobs}
              prefix={<FileTextOutlined />}
              suffix={`/ ${stats.completedJobs} 完成`}
              valueStyle={{ color: '#722ed1' }}
            />
          </Card>
        </Col>
        <Col xs={24} sm={12} md={6}>
          <Card>
            <Statistic
              title="用户数量"
              value={stats.totalUsers}
              prefix={<UserOutlined />}
              suffix={`/ ${stats.activeUsers} 活跃`}
              valueStyle={{ color: '#fa8c16' }}
            />
          </Card>
        </Col>
      </Row>

      {/* 系统状态提醒 */}
      <Row gutter={16} style={{ marginBottom: 24 }}>
        <Col span={24}>
          <Alert
            message="系统状态正常"
            description="所有核心服务运行正常，4台设备需要补充耗材。"
            type="info"
            showIcon
            closable
          />
        </Col>
      </Row>

      {/* 图表和设备状态 */}
      <Row gutter={16} style={{ marginBottom: 24 }}>
        <Col xs={24} lg={12}>
          <Card title="打印任务趋势" style={{ height: 400 }}>
            <div id="printJobsChart" style={{ width: '100%', height: 300 }}></div>
          </Card>
        </Col>
        <Col xs={24} lg={12}>
          <Card title="设备状态概览" style={{ height: 400 }}>
            <Space direction="vertical" style={{ width: '100%' }}>
              <div>
                <CheckCircleOutlined style={{ color: '#52c41a', marginRight: 8 }} />
                在线设备: {stats.onlinePrinters} / {stats.totalPrinters}
              </div>
              <div>
                <ExclamationCircleOutlined style={{ color: '#faad14', marginRight: 8 }} />
                需要维护: 2 台设备
              </div>
              <div>
                <StopOutlined style={{ color: '#ff4d4f', marginRight: 8 }} />
                离线设备: {stats.totalPrinters - stats.onlinePrinters} 台
              </div>
              <div style={{ marginTop: 16 }}>
                <h4>设备健康度</h4>
                <Progress percent={85} status="active" />
              </div>
              <div style={{ marginTop: 16 }}>
                <h4>网络连接状态</h4>
                <Progress percent={92} />
              </div>
            </Space>
          </Card>
        </Col>
      </Row>

      {/* 设备状态表格 */}
      <Row gutter={16} style={{ marginBottom: 24 }}>
        <Col span={24}>
          <Card title="打印机状态">
            <Table
              columns={printerColumns}
              dataSource={printers}
              pagination={false}
              size="small"
            />
          </Card>
        </Col>
      </Row>

      {/* 最近打印任务 */}
      <Row gutter={16}>
        <Col span={24}>
          <Card title="最近打印任务">
            <Table
              columns={jobColumns}
              dataSource={printJobs}
              pagination={{ pageSize: 5, showSizeChanger: false }}
              size="small"
            />
          </Card>
        </Col>
      </Row>
    </div>
  );
};

export default Dashboard;
