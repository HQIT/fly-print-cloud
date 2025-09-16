import React, { useState, useEffect } from 'react';
import { Card, Table, Tag, Space, Row, Col, Statistic, Select, DatePicker, Button, message } from 'antd';
import { 
  CheckCircleOutlined,
  ExclamationCircleOutlined,
  ClockCircleOutlined,
  PlayCircleOutlined,
  FileTextOutlined,
  ReloadOutlined,
  StopOutlined
} from '@ant-design/icons';

const { RangePicker } = DatePicker;
const { Option } = Select;

// 打印任务接口（适配后端数据模型）
interface PrintJob {
  id: string;
  name: string; // 后端字段名
  user_name: string; // 后端字段名
  printer_id: string; // 后端字段名
  status: 'pending' | 'printing' | 'completed' | 'failed' | 'cancelled';
  created_at: string; // 后端字段名
  page_count: number; // 后端字段名
  file_path: string;
  file_size: number;
  copies: number;
  priority: number;
  key?: string;
}

// Print Jobs 服务类
class PrintJobsService {
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

  async getPrintJobs(page = 1, pageSize = 10, status?: string): Promise<{ jobs: PrintJob[]; total: number }> {
    try {
      const token = await this.getToken();
      let url = `/api/v1/admin/print-jobs?page=${page}&pageSize=${pageSize}`;
      if (status) {
        url += `&status=${status}`;
      }
      
      const response = await fetch(url, {
        headers: {
          ...(token && { 'Authorization': `Bearer ${token}` }),
        },
      });
      
      if (response.ok) {
        const result = await response.json();
        return result.data;
      }
    } catch (error) {
      console.error('获取打印任务列表失败:', error);
    }
    
    // 返回模拟数据作为fallback（适配后端格式）
    return {
      jobs: [
        {
          id: 'JOB-2024-001',
          name: '合同文件.pdf',
          user_name: '张三',
          printer_id: 'printer-1',
          status: 'completed',
          created_at: '2024-01-15T10:30:00Z',
          page_count: 5,
          file_path: '/uploads/contract.pdf',
          file_size: 1024000,
          copies: 1,
          priority: 5,
        },
        {
          id: 'JOB-2024-002',
          name: '报告.docx',
          user_name: '李四',
          printer_id: 'printer-2',
          status: 'printing',
          created_at: '2024-01-15T11:15:00Z',
          page_count: 12,
          file_path: '/uploads/report.docx',
          file_size: 2048000,
          copies: 2,
          priority: 7,
        },
        {
          id: 'JOB-2024-003',
          name: '图表分析.xlsx',
          user_name: '王五',
          printer_id: 'printer-3',
          status: 'pending',
          created_at: '2024-01-15T11:45:00Z',
          page_count: 3,
          file_path: '/uploads/analysis.xlsx',
          file_size: 512000,
          copies: 1,
          priority: 3,
        },
        {
          id: 'JOB-2024-004',
          name: '项目计划.pptx',
          user_name: '赵六',
          printer_id: 'printer-4',
          status: 'failed',
          created_at: '2024-01-15T12:00:00Z',
          page_count: 20,
          file_path: '/uploads/project.pptx',
          file_size: 4096000,
          copies: 3,
          priority: 8,
        },
        {
          id: 'JOB-2024-005',
          name: '财务报表.pdf',
          user_name: '钱七',
          printer_id: 'printer-5',
          status: 'completed',
          created_at: '2024-01-15T09:30:00Z',
          page_count: 8,
          file_path: '/uploads/finance.pdf',
          file_size: 1536000,
          copies: 1,
          priority: 6,
        },
        {
          id: 'JOB-2024-006',
          name: '会议纪要.docx',
          user_name: '孙八',
          printer_id: 'printer-6',
          status: 'pending',
          created_at: '2024-01-15T14:20:00Z',
          page_count: 2,
          file_path: '/uploads/meeting.docx',
          file_size: 256000,
          copies: 1,
          priority: 4,
        },
        {
          id: 'JOB-2024-007',
          name: '设计图纸.pdf',
          user_name: '周九',
          printer_id: 'printer-1',
          status: 'failed',
          created_at: '2024-01-15T13:45:00Z',
          page_count: 15,
          file_path: '/uploads/design.pdf',
          file_size: 8192000,
          copies: 1,
          priority: 9,
        },
        {
          id: 'JOB-2024-008',
          name: '产品手册.pdf',
          user_name: '吴十',
          printer_id: 'printer-2',
          status: 'completed',
          created_at: '2024-01-15T08:15:00Z',
          page_count: 25,
          file_path: '/uploads/manual.pdf',
          file_size: 6144000,
          copies: 2,
          priority: 5,
        },
      ],
      total: 8,
    };
  }
}

const printJobsService = new PrintJobsService();

// Print Jobs 组件
const PrintJobs: React.FC = () => {
  const [printJobs, setPrintJobs] = useState<PrintJob[]>([]);
  const [loading, setLoading] = useState(true);
  const [statusFilter, setStatusFilter] = useState<string>('');
  const [currentPage, setCurrentPage] = useState(1);
  const [pageSize, setPageSize] = useState(10);
  const [total, setTotal] = useState(0);

  // 加载打印任务数据
  const loadPrintJobs = async (page = 1, size = 10, status = '') => {
    try {
      setLoading(true);
      const result = await printJobsService.getPrintJobs(page, size, status);
      setPrintJobs(result.jobs.map(job => ({ ...job, key: job.id })));
      setTotal(result.total);
    } catch (error) {
      console.error('加载打印任务失败:', error);
      // 设置 fallback 数据
        const fallbackJobs = [
          {
            id: 'JOB-2024-001',
            name: '合同文件.pdf',
            user_name: '张三',
            printer_id: 'printer-1',
            status: 'completed' as const,
            created_at: '2024-01-15T10:30:00Z',
            page_count: 5,
            file_path: '/uploads/contract.pdf',
            file_size: 1024000,
            copies: 1,
            priority: 5,
          },
          {
            id: 'JOB-2024-002',
            name: '报告.docx',
            user_name: '李四',
            printer_id: 'printer-2',
            status: 'printing' as const,
            created_at: '2024-01-15T11:15:00Z',
            page_count: 12,
            file_path: '/uploads/report.docx',
            file_size: 2048000,
            copies: 2,
            priority: 7,
          },
          {
            id: 'JOB-2024-003',
            name: '图表分析.xlsx',
            user_name: '王五',
            printer_id: 'printer-3',
            status: 'pending' as const,
            created_at: '2024-01-15T11:45:00Z',
            page_count: 3,
            file_path: '/uploads/analysis.xlsx',
            file_size: 512000,
            copies: 1,
            priority: 3,
          },
          {
            id: 'JOB-2024-004',
            name: '项目计划.pptx',
            user_name: '赵六',
            printer_id: 'printer-4',
            status: 'failed' as const,
            created_at: '2024-01-15T12:00:00Z',
            page_count: 20,
            file_path: '/uploads/project.pptx',
            file_size: 4096000,
            copies: 3,
            priority: 8,
          },
        ];
      setPrintJobs(fallbackJobs.map(job => ({ ...job, key: job.id })));
      setTotal(fallbackJobs.length);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    loadPrintJobs(currentPage, pageSize, statusFilter);
  }, [currentPage, pageSize, statusFilter]);

  const getStatusColor = (status: string) => {
    switch (status) {
      case 'completed': return 'success';
      case 'printing': return 'processing';
      case 'pending': return 'warning';
      case 'failed': return 'error';
      default: return 'default';
    }
  };

  const getStatusText = (status: string) => {
    switch (status) {
      case 'completed': return '已完成';
      case 'printing': return '打印中';
      case 'pending': return '等待中';
      case 'failed': return '失败';
      default: return '未知';
    }
  };

  const getStatusIcon = (status: string) => {
    switch (status) {
      case 'completed': return <CheckCircleOutlined />;
      case 'printing': return <PlayCircleOutlined />;
      case 'pending': return <ClockCircleOutlined />;
      case 'failed': return <ExclamationCircleOutlined />;
      default: return <ClockCircleOutlined />;
    }
  };

  const handleStatusFilterChange = (value: string) => {
    setStatusFilter(value);
    setCurrentPage(1);
  };

  const handleRefresh = () => {
    loadPrintJobs(currentPage, pageSize, statusFilter);
  };

  const columns = [
    {
      title: '任务ID',
      dataIndex: 'id',
      key: 'id',
      width: 130,
      render: (text: string) => <code>{text}</code>,
    },
    {
      title: '文件名',
      dataIndex: 'name',
      key: 'name',
      width: 200,
      ellipsis: true,
      render: (text: string) => <strong>{text}</strong>,
    },
    {
      title: '用户',
      dataIndex: 'user_name',
      key: 'user_name',
      width: 100,
    },
    {
      title: '打印机ID',
      dataIndex: 'printer_id',
      key: 'printer_id',
      width: 120,
      ellipsis: true,
      render: (text: string) => <code>{text}</code>,
    },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      width: 120,
      render: (status: string) => (
        <Tag color={getStatusColor(status)} icon={getStatusIcon(status)}>
          {getStatusText(status)}
        </Tag>
      ),
    },
    {
      title: '页数',
      dataIndex: 'page_count',
      key: 'page_count',
      width: 80,
      render: (pages: number) => `${pages} 页`,
    },
    {
      title: '创建时间',
      dataIndex: 'created_at',
      key: 'created_at',
      width: 150,
      render: (timestamp: string) => {
        const date = new Date(timestamp);
        return date.toLocaleString('zh-CN');
      },
    },
    {
      title: '操作',
      key: 'action',
      width: 150,
      render: (record: PrintJob) => (
        <Space>
          <a onClick={() => message.info(`查看任务 ${record.id} 详情`)}>详情</a>
          {record.status === 'pending' && (
            <a onClick={() => message.info(`取消任务 ${record.id}`)}>取消</a>
          )}
          {record.status === 'failed' && (
            <a onClick={() => message.info(`重试任务 ${record.id}`)}>重试</a>
          )}
        </Space>
      ),
    },
  ];

  return (
    <div>
      <div style={{ marginBottom: 16, display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
        <h2 style={{ margin: 0 }}>打印任务管理</h2>
        <Space>
          <Select
            placeholder="筛选状态"
            style={{ width: 120 }}
            allowClear
            value={statusFilter || undefined}
            onChange={handleStatusFilterChange}
          >
            <Option value="pending">等待中</Option>
            <Option value="printing">打印中</Option>
            <Option value="completed">已完成</Option>
            <Option value="failed">失败</Option>
          </Select>
          <Button icon={<ReloadOutlined />} onClick={handleRefresh}>
            刷新
          </Button>
        </Space>
      </div>

      <Card>
        <Table
          columns={columns}
          dataSource={printJobs}
          loading={loading}
          pagination={{
            current: currentPage,
            pageSize: pageSize,
            total: total,
            showSizeChanger: true,
            showQuickJumper: true,
            showTotal: (total) => `共 ${total} 个任务`,
            onChange: (page, size) => {
              setCurrentPage(page);
              setPageSize(size || 10);
            },
          }}
          scroll={{ x: 1000 }}
        />
      </Card>

      {/* 统计信息 */}
      <Row gutter={16} style={{ marginTop: 16 }}>
        <Col span={6}>
          <Card>
            <Statistic
              title="总任务数"
              value={printJobs.length}
              prefix={<FileTextOutlined />}
            />
          </Card>
        </Col>
        <Col span={6}>
          <Card>
            <Statistic
              title="已完成"
              value={printJobs.filter(job => job.status === 'completed').length}
              prefix={<CheckCircleOutlined style={{ color: '#52c41a' }} />}
              valueStyle={{ color: '#52c41a' }}
            />
          </Card>
        </Col>
        <Col span={6}>
          <Card>
            <Statistic
              title="进行中"
              value={printJobs.filter(job => job.status === 'printing').length}
              prefix={<PlayCircleOutlined style={{ color: '#1890ff' }} />}
              valueStyle={{ color: '#1890ff' }}
            />
          </Card>
        </Col>
        <Col span={6}>
          <Card>
            <Statistic
              title="失败任务"
              value={printJobs.filter(job => job.status === 'failed').length}
              prefix={<ExclamationCircleOutlined style={{ color: '#ff4d4f' }} />}
              valueStyle={{ color: '#ff4d4f' }}
            />
          </Card>
        </Col>
      </Row>
    </div>
  );
};

export default PrintJobs;
