import React, { useState, useEffect } from 'react';
import { Card, Table, Tag, Space, Button, Select, Input, message, Popconfirm, Modal, Form, InputNumber, Radio } from 'antd';
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

// Edge Node 接口
interface EdgeNode {
  id: string;
  name: string;
  display_name: string;
  status: string;
}

// Printer 接口
interface Printer {
  id: string;
  name: string;
  display_name: string;
  status: string;
  edge_node_id: string;
  capabilities: {
    paper_sizes: string[];
    color_support: boolean;
    duplex_support: boolean;
    resolution: string;
    print_speed: string;
    media_types: string[];
  };
}

// 重新打印表单数据
interface ReprintFormData {
  edge_node_id: string;
  printer_id: string;
  copies: number;
  color_mode: 'color' | 'grayscale';
  duplex_mode: 'single' | 'duplex';
  paper_size: string;
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
  
  // 重新打印Modal相关状态
  const [reprintModalVisible, setReprintModalVisible] = useState(false);
  const [reprintJob, setReprintJob] = useState<PrintJob | null>(null);
  const [edgeNodes, setEdgeNodes] = useState<EdgeNode[]>([]);
  const [printers, setPrinters] = useState<Printer[]>([]);
  const [selectedEdgeNodeId, setSelectedEdgeNodeId] = useState<string>('');
  const [selectedPrinter, setSelectedPrinter] = useState<Printer | null>(null);
  const [form] = Form.useForm();

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

  // 打开重新打印Modal
  const handleReprintJob = async (job: PrintJob) => {
    setReprintJob(job);
    setReprintModalVisible(true);
    
    // 加载Edge Nodes
    await loadEdgeNodes();
    
    // 预设表单默认值
    form.setFieldsValue({
      edge_node_id: '', // 需要用户选择
      printer_id: job.printer_id,
      copies: job.copies || 1,
      color_mode: job.color_mode || 'grayscale',
      duplex_mode: job.duplex_mode || 'single',
      paper_size: job.paper_size || 'A4',
    });
  };

  // 加载Edge Nodes
  const loadEdgeNodes = async () => {
    try {
      const token = await printJobsService.getToken();
      const response = await fetch('/api/v1/admin/edge-nodes', {
        headers: {
          ...(token && { 'Authorization': `Bearer ${token}` }),
        },
      });
      
      if (response.ok) {
        const result = await response.json();
        setEdgeNodes(result.data?.items || []);
      }
    } catch (error) {
      console.error('加载Edge Nodes失败:', error);
    }
  };

  // 加载指定Edge Node的打印机
  const loadPrinters = async (edgeNodeId: string) => {
    try {
      const token = await printJobsService.getToken();
      const response = await fetch(`/api/v1/admin/printers?edge_node_id=${edgeNodeId}`, {
        headers: {
          ...(token && { 'Authorization': `Bearer ${token}` }),
        },
      });
      
      if (response.ok) {
        const result = await response.json();
        setPrinters(result.data?.items || []);
      }
    } catch (error) {
      console.error('加载打印机失败:', error);
    }
  };

  // Edge Node选择变化
  const handleEdgeNodeChange = (edgeNodeId: string) => {
    setSelectedEdgeNodeId(edgeNodeId);
    setSelectedPrinter(null); // 清空选中的打印机
    form.setFieldsValue({ printer_id: '' }); // 清空打印机选择
    loadPrinters(edgeNodeId);
  };

  // 打印机选择变化
  const handlePrinterChange = (printerId: string) => {
    const printer = printers.find(p => p.id === printerId);
    setSelectedPrinter(printer || null);
    
    // 确保表单中的printer_id正确设置
    form.setFieldsValue({ printer_id: printerId });
    
    // 根据打印机能力调整表单值
    if (printer) {
      const formValues = form.getFieldsValue();
      const updates: any = {};
      
      // 如果打印机不支持彩色，强制设为黑白
      if (!printer.capabilities.color_support && formValues.color_mode === 'color') {
        updates.color_mode = 'grayscale';
      }
      
      // 如果打印机不支持双面，强制设为单面
      if (!printer.capabilities.duplex_support && formValues.duplex_mode === 'duplex') {
        updates.duplex_mode = 'single';
      }
      
      // 如果当前纸张大小不支持，设为第一个支持的纸张大小
      if (printer.capabilities.paper_sizes.length > 0 && 
          !printer.capabilities.paper_sizes.includes(formValues.paper_size)) {
        updates.paper_size = printer.capabilities.paper_sizes[0];
      }
      
      // 批量更新表单值
      if (Object.keys(updates).length > 0) {
        form.setFieldsValue(updates);
      }
    }
  };

  // 提交重新打印
  const handleReprintSubmit = async (values: ReprintFormData) => {
    if (!reprintJob) return;

    try {
      const token = await printJobsService.getToken();
      const response = await fetch(`/api/v1/admin/print-jobs/${reprintJob.id}/reprint`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          ...(token && { 'Authorization': `Bearer ${token}` }),
        },
        body: JSON.stringify({
          printer_id: values.printer_id,
          copies: values.copies,
          paper_size: values.paper_size,
          color_mode: values.color_mode,
          duplex_mode: values.duplex_mode,
        }),
      });
      
      if (response.ok) {
        message.success('重新打印任务创建成功');
        setReprintModalVisible(false);
        setSelectedPrinter(null);
        setSelectedEdgeNodeId('');
        form.resetFields();
        loadPrintJobs(currentPage, pageSize, statusFilter);
      } else {
        const result = await response.json();
        message.error(result.error || '重新打印任务创建失败');
      }
    } catch (error) {
      console.error('重新打印任务失败:', error);
      message.error('重新打印任务失败');
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
          
          {/* 重新打印 - 已完成、失败、取消的任务都可以重新打印 */}
          {(record.status === 'completed' || record.status === 'failed' || record.status === 'cancelled') && (
            <Button size="small" type="link" onClick={() => handleReprintJob(record)}>
              重新打印
            </Button>
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

      {/* 重新打印Modal */}
      <Modal
        title={`重新打印 - ${reprintJob?.name}`}
        open={reprintModalVisible}
        onCancel={() => {
          setReprintModalVisible(false);
          setSelectedPrinter(null);
          setSelectedEdgeNodeId('');
          form.resetFields();
        }}
        footer={null}
        width={600}
      >
        <Form
          form={form}
          layout="vertical"
          onFinish={handleReprintSubmit}
        >
          <Form.Item
            name="edge_node_id"
            label="选择Edge Node"
            rules={[{ required: true, message: '请选择Edge Node' }]}
          >
            <Select
              placeholder="请选择Edge Node"
              onChange={handleEdgeNodeChange}
              showSearch
              optionFilterProp="children"
            >
              {edgeNodes.map(node => (
                <Select.Option key={node.id} value={node.id}>
                  {node.display_name || node.name} ({node.status})
                </Select.Option>
              ))}
            </Select>
          </Form.Item>

          <Form.Item
            name="printer_id"
            label="选择打印机"
            rules={[{ required: true, message: '请选择打印机' }]}
          >
            <Select
              placeholder="请先选择Edge Node"
              disabled={!selectedEdgeNodeId}
              showSearch
              optionFilterProp="children"
              onChange={handlePrinterChange}
            >
              {printers.map(printer => (
                <Select.Option key={printer.id} value={printer.id}>
                  {printer.display_name || printer.name} ({printer.status})
                </Select.Option>
              ))}
            </Select>
            {selectedPrinter && (
              <div style={{ fontSize: '12px', color: '#666', marginTop: '4px' }}>
                <strong>打印机能力：</strong>
                {selectedPrinter.capabilities.color_support ? '支持彩色' : '仅黑白'}
                {', '}
                {selectedPrinter.capabilities.duplex_support ? '支持双面' : '仅单面'}
                {selectedPrinter.capabilities.paper_sizes.length > 0 && (
                  <>，支持纸张：{selectedPrinter.capabilities.paper_sizes.join(', ')}</>
                )}
              </div>
            )}
          </Form.Item>

          <Form.Item
            name="copies"
            label="打印份数"
            rules={[{ required: true, message: '请输入打印份数' }]}
          >
            <InputNumber min={1} max={99} />
          </Form.Item>

          <Form.Item
            name="paper_size"
            label="纸张大小"
            rules={[{ required: true, message: '请选择纸张大小' }]}
          >
            <Select disabled={!selectedPrinter}>
              {selectedPrinter?.capabilities.paper_sizes.length > 0 ? (
                selectedPrinter.capabilities.paper_sizes.map(size => (
                  <Select.Option key={size} value={size}>{size}</Select.Option>
                ))
              ) : (
                // 默认选项（如果打印机没有指定支持的纸张大小）
                <>
                  <Select.Option value="A4">A4</Select.Option>
                  <Select.Option value="A3">A3</Select.Option>
                  <Select.Option value="Letter">Letter</Select.Option>
                  <Select.Option value="Legal">Legal</Select.Option>
                </>
              )}
            </Select>
          </Form.Item>

          <Form.Item
            name="color_mode"
            label="颜色模式"
            rules={[{ required: true, message: '请选择颜色模式' }]}
          >
            <Radio.Group disabled={!selectedPrinter}>
              <Radio value="grayscale">黑白</Radio>
              {selectedPrinter?.capabilities.color_support && (
                <Radio value="color">彩色</Radio>
              )}
            </Radio.Group>
            {selectedPrinter && !selectedPrinter.capabilities.color_support && (
              <div style={{ fontSize: '12px', color: '#999', marginTop: '4px' }}>
                该打印机不支持彩色打印
              </div>
            )}
          </Form.Item>

          <Form.Item
            name="duplex_mode"
            label="双面模式"
            rules={[{ required: true, message: '请选择双面模式' }]}
          >
            <Radio.Group disabled={!selectedPrinter}>
              <Radio value="single">单面</Radio>
              {selectedPrinter?.capabilities.duplex_support && (
                <Radio value="duplex">双面</Radio>
              )}
            </Radio.Group>
            {selectedPrinter && !selectedPrinter.capabilities.duplex_support && (
              <div style={{ fontSize: '12px', color: '#999', marginTop: '4px' }}>
                该打印机不支持双面打印
              </div>
            )}
          </Form.Item>

          <Form.Item style={{ marginBottom: 0, textAlign: 'right' }}>
            <Space>
              <Button onClick={() => setReprintModalVisible(false)}>
                取消
              </Button>
              <Button type="primary" htmlType="submit">
                确认打印
              </Button>
            </Space>
          </Form.Item>
        </Form>
      </Modal>
    </div>
  );
};

export default PrintJobs;