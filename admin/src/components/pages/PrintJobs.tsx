import React, { useState, useEffect } from 'react';
import { Card, Table, Tag, Space, Button, Select, Input, message, Popconfirm } from 'antd';
import { 
  ReloadOutlined,
  SearchOutlined,
  DeleteOutlined,
  FileTextOutlined,
  UserOutlined,
  PrinterOutlined
} from '@ant-design/icons';

// 打印任务接口定义
interface PrintJob {
  id: string;
  name: string;
  user_name: string;
  printer_id: string;
  status: 'pending' | 'dispatched' | 'downloading' | 'printing' | 'completed' | 'failed' | 'cancelled';
  created_at: string;
  updated_at: string;
  page_count: number;
  file_path: string;
  file_url: string;
  file_size: number;
  copies: number;
  paper_size: string;
  color_mode: string;
  duplex_mode: string;
  start_time: string;
  end_time: string;
  error_message: string;
  retry_count: number;
  max_retries: number;
  key?: string;
}

// Print Jobs 服务类
class PrintJobsService {
  async getToken(): Promise<string | null> {
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

  async getPrintJobs(page = 1, pageSize = 10, status = ''): Promise<{ jobs: PrintJob[]; total: number; page: number; pageSize: number }> {
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
        return {
          jobs: result?.jobs || [],
          total: result?.pagination?.total || result?.jobs?.length || 0,
          page: page,
          pageSize: pageSize
        };
      }
    } catch (error) {
      console.error('获取打印任务列表失败:', error);
    }
    
    // API调用失败时返回空数据
    return {
      jobs: [],
      total: 0,
      page: page,
      pageSize: pageSize
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
      setCurrentPage(page);
      setPageSize(size);
    } catch (error) {
      console.error('加载打印任务失败:', error);
      message.error('加载打印任务失败，请稍后重试');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    loadPrintJobs(currentPage, pageSize, statusFilter);
  }, []);

  // 状态标签映射
  const getStatusTag = (status: string) => {
    switch (status) {
      case 'pending':
        return <Tag color="default">等待中</Tag>;
      case 'dispatched':
        return <Tag color="blue">已分发</Tag>;
      case 'downloading':
        return <Tag color="cyan">下载中</Tag>;
      case 'printing':
        return <Tag color="processing">打印中</Tag>;
      case 'completed':
        return <Tag color="success">已完成</Tag>;
      case 'failed':
        return <Tag color="error">失败</Tag>;
      case 'cancelled':
        return <Tag color="default">已取消</Tag>;
      default:
        return <Tag color="default">{status}</Tag>;
    }
  };

  // 处理状态筛选
  const handleStatusChange = (value: string) => {
    setStatusFilter(value);
    setCurrentPage(1);
    loadPrintJobs(1, pageSize, value);
  };

  // 处理分页变化
  const handleTableChange = (page: number, size?: number) => {
    const newSize = size || pageSize;
    setCurrentPage(page);
    setPageSize(newSize);
    loadPrintJobs(page, newSize, statusFilter);
  };

  // 刷新数据
  const handleRefresh = () => {
    loadPrintJobs(currentPage, pageSize, statusFilter);
  };

  // 删除任务
  const handleDeleteJob = async (jobId: string) => {
    try {
      const token = await printJobsService.getToken();
      const response = await fetch(`/api/v1/admin/print-jobs/${jobId}`, {
        method: 'DELETE',
        headers: {
          ...(token && { 'Authorization': `Bearer ${token}` }),
        },
      });
      
      if (response.ok) {
        message.success('删除任务成功');
        loadPrintJobs(currentPage, pageSize, statusFilter);
      } else {
        message.error('删除任务失败');
      }
    } catch (error) {
      console.error('删除任务失败:', error);
      message.error('删除任务失败');
    }
  };

  // 取消任务
  const handleCancelJob = async (jobId: string) => {
    try {
      const token = await printJobsService.getToken();
      const response = await fetch(`/api/v1/admin/print-jobs/${jobId}`, {
        method: 'PUT',
        headers: {
          'Content-Type': 'application/json',
          ...(token && { 'Authorization': `Bearer ${token}` }),
        },
        body: JSON.stringify({ status: 'cancelled' }),
      });
      
      if (response.ok) {
        message.success('取消任务成功');
        loadPrintJobs(currentPage, pageSize, statusFilter);
      } else {
        message.error('取消任务失败');
      }
    } catch (error) {
      console.error('取消任务失败:', error);
      message.error('取消任务失败');
    }
  };

  // 表格列定义
  const columns = [
    {
      title: '任务名称',
      dataIndex: 'name',
      key: 'name',
      width: 200,
      render: (text: string) => (
        <Space>
          <FileTextOutlined />
          <span title={text}>
            {text && text.length > 30 ? `${text.substring(0, 30)}...` : text}
          </span>
        </Space>
      ),
    },
    {
      title: '用户',
      dataIndex: 'user_name',
      key: 'user_name',
      render: (text: string) => (
        <Space>
          <UserOutlined />
          {text || '-'}
        </Space>
      ),
    },
    {
      title: '打印机ID',
      dataIndex: 'printer_id',
      key: 'printer_id',
      render: (text: string) => (
        <Space>
          <PrinterOutlined />
          {text ? text.substring(0, 8) + '...' : '-'}
        </Space>
      ),
    },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      render: (status: string) => getStatusTag(status),
    },
    {
      title: '页数',
      dataIndex: 'page_count',
      key: 'page_count',
      render: (count: number) => count || '-',
    },
    {
      title: '份数',
      dataIndex: 'copies',
      key: 'copies',
      render: (copies: number) => copies || 1,
    },
    {
      title: '创建时间',
      dataIndex: 'created_at',
      key: 'created_at',
      render: (time: string) => {
        if (!time) return '-';
        const date = new Date(time);
        return date.toLocaleString('zh-CN');
      },
    },
    {
      title: '文件大小',
      dataIndex: 'file_size',
      key: 'file_size',
      render: (size: number) => {
        if (!size) return '-';
        if (size < 1024) return `${size} B`;
        if (size < 1024 * 1024) return `${(size / 1024).toFixed(1)} KB`;
        return `${(size / (1024 * 1024)).toFixed(1)} MB`;
      },
    },
    {
      title: '操作',
      key: 'action',
      width: 150,
      render: (_, record: PrintJob) => (
        <Space size="small">
          {/* 取消任务 - 只有pending和dispatched状态可以取消 */}
          {(record.status === 'pending' || record.status === 'dispatched') && (
            <Popconfirm
              title="确定要取消这个任务吗？"
              onConfirm={() => handleCancelJob(record.id)}
              okText="确定"
              cancelText="取消"
            >
              <Button size="small" type="link">
                取消
              </Button>
            </Popconfirm>
          )}
          
          {/* 删除任务 - 只有completed、failed、cancelled状态可以删除 */}
          {(record.status === 'completed' || record.status === 'failed' || record.status === 'cancelled') && (
            <Popconfirm
              title="确定要删除这个任务吗？"
              onConfirm={() => handleDeleteJob(record.id)}
              okText="确定"
              cancelText="取消"
            >
              <Button size="small" type="link" danger icon={<DeleteOutlined />}>
                删除
              </Button>
            </Popconfirm>
          )}
        </Space>
      ),
    },
  ];

  return (
    <div style={{ padding: '24px' }}>
      <h2>打印任务管理</h2>
      
      <Card>
        {/* 筛选和操作栏 */}
        <div style={{ marginBottom: '16px', display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
          <Space>
            <Select
              placeholder="选择状态筛选"
              allowClear
              style={{ width: 150 }}
              value={statusFilter || undefined}
              onChange={handleStatusChange}
            >
              <Select.Option value="pending">等待中</Select.Option>
              <Select.Option value="dispatched">已分发</Select.Option>
              <Select.Option value="downloading">下载中</Select.Option>
              <Select.Option value="printing">打印中</Select.Option>
              <Select.Option value="completed">已完成</Select.Option>
              <Select.Option value="failed">失败</Select.Option>
              <Select.Option value="cancelled">已取消</Select.Option>
            </Select>
          </Space>
          
          <Space>
            <Button 
              icon={<ReloadOutlined />} 
              onClick={handleRefresh}
              loading={loading}
            >
              刷新
            </Button>
          </Space>
        </div>

        {/* 打印任务表格 */}
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
            showTotal: (total, range) =>
              `第 ${range[0]}-${range[1]} 条，共 ${total} 条`,
            onChange: handleTableChange,
            onShowSizeChange: handleTableChange,
            pageSizeOptions: ['10', '20', '50', '100'],
          }}
          size="middle"
          scroll={{ x: 1200 }}
        />
      </Card>
    </div>
  );
};

export default PrintJobs;