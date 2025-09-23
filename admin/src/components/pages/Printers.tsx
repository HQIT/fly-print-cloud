import React, { useState, useEffect } from 'react';
import { Card, Table, Tag, Space, Row, Col, Statistic, Progress, message, Select, Button, Popconfirm } from 'antd';
import { 
  CheckCircleOutlined,
  ExclamationCircleOutlined,
  StopOutlined,
  PrinterOutlined,
  PlayCircleOutlined,
  DeleteOutlined
} from '@ant-design/icons';

// 打印机接口（适配后端数据模型）
interface PrinterStatus {
  id: string;
  name: string;
  model: string;
  location?: string; // 后端可能为空
  status: 'ready' | 'printing' | 'error' | 'offline'; // 后端状态值
  edge_node_id: string;
  edge_node_name?: string; // Edge Node 名称
  queue_length: number;
  key?: string;
}

// Edge Node 接口
interface EdgeNode {
  id: string;
  name: string;
}

// Printers 服务类
class PrintersService {
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

  async getEdgeNodes(): Promise<EdgeNode[]> {
    try {
      const token = await this.getToken();
      const response = await fetch('/api/v1/admin/edge-nodes', {
        headers: {
          ...(token && { 'Authorization': `Bearer ${token}` }),
        },
      });
      
      if (response.ok) {
        const result = await response.json();
        return result.data.items || [];
      }
    } catch (error) {
      console.error('获取边缘节点列表失败:', error);
    }
    
    return [];
  }

  async getPrintersWithEdgeNodes(): Promise<{ printers: PrinterStatus[], edgeNodes: EdgeNode[] }> {
    try {
      const token = await this.getToken();
      
      // 同时获取打印机和边缘节点数据
      const [printersResponse, edgeNodesResponse] = await Promise.all([
        fetch('/api/v1/admin/printers', {
          headers: { ...(token && { 'Authorization': `Bearer ${token}` }) },
        }),
        fetch('/api/v1/admin/edge-nodes', {
          headers: { ...(token && { 'Authorization': `Bearer ${token}` }) },
        })
      ]);
      
      const printers = printersResponse.ok ? (await printersResponse.json()).data.items || [] : [];
      const edgeNodes = edgeNodesResponse.ok ? (await edgeNodesResponse.json()).data.items || [] : [];
      
      // 创建 Edge Node 映射
      const edgeNodeMap: { [key: string]: string } = {};
      edgeNodes.forEach((node: EdgeNode) => {
        edgeNodeMap[node.id] = node.name;
      });
      
      // 合并数据
      const printersWithEdgeNode = printers.map((printer: any) => ({
        ...printer,
        edge_node_name: edgeNodeMap[printer.edge_node_id] || printer.edge_node_id
      }));
      
      return { printers: printersWithEdgeNode, edgeNodes };
    } catch (error) {
      console.error('获取数据失败:', error);
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
}

const printersService = new PrintersService();

// Printers 组件
const Printers: React.FC = () => {
  const [printers, setPrinters] = useState<PrinterStatus[]>([]);
  const [edgeNodes, setEdgeNodes] = useState<EdgeNode[]>([]);
  const [filteredPrinters, setFilteredPrinters] = useState<PrinterStatus[]>([]);
  const [selectedEdgeNode, setSelectedEdgeNode] = useState<string>('');
  const [loading, setLoading] = useState(true);
  const [currentPage, setCurrentPage] = useState(1);
  const [pageSize, setPageSize] = useState(10);

  // 获取token
  const getToken = async (): Promise<string | null> => {
    try {
      const response = await fetch('/auth/me');
      const result = await response.json();
      
      if (result.code === 200 && result.data.access_token) {
        return result.data.access_token;
      }
      return null;
    } catch (error) {
      console.error('获取token失败:', error);
      return null;
    }
  };

  // 删除打印机
  const handleDeletePrinter = async (printerId: string, printerName: string) => {
    try {
      const token = await getToken();
      if (!token) {
        message.error('未找到认证令牌，请重新登录');
        return;
      }

      const response = await fetch(`/api/v1/admin/printers/${printerId}`, {
        method: 'DELETE',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${token}`,
        },
      });

      if (response.ok) {
        message.success(`打印机 ${printerName} 删除成功`);
        loadData(); // 重新加载数据
      } else {
        const error = await response.json();
        message.error(`删除失败: ${error.message || '未知错误'}`);
      }
    } catch (error) {
      console.error('删除打印机失败:', error);
      message.error('删除失败，请稍后重试');
    }
  };

  // 加载数据函数
  const loadData = async () => {
    try {
      setLoading(true);
      const { printers: printerList, edgeNodes: edgeNodeList } = await printersService.getPrintersWithEdgeNodes();
      
      const printersWithKey = printerList.map(printer => ({ ...printer, key: printer.id }));
      setPrinters(printersWithKey);
      setEdgeNodes(edgeNodeList);
      setFilteredPrinters(printersWithKey);
    } catch (error) {
      console.error('加载数据失败:', error);
      message.error('加载打印机数据失败，请稍后重试');
      // 设置为空数组
      setPrinters([]);
      setEdgeNodes([]);
      setFilteredPrinters([]);
    } finally {
      setLoading(false);
    }
  };

  // 初始加载数据
  useEffect(() => {
    loadData();
  }, []);

  // Edge Node 筛选逻辑
  useEffect(() => {
    if (selectedEdgeNode === '') {
      setFilteredPrinters(printers);
    } else {
      setFilteredPrinters(printers.filter(printer => printer.edge_node_id === selectedEdgeNode));
    }
  }, [selectedEdgeNode, printers]);

  const getStatusColor = (status: string) => {
    switch (status) {
      case 'ready': return 'success';
      case 'printing': return 'processing';
      case 'offline': return 'default';
      case 'error': return 'error';
      default: return 'default';
    }
  };

  const getStatusText = (status: string) => {
    switch (status) {
      case 'ready': return '就绪';
      case 'printing': return '打印中';
      case 'offline': return '离线';
      case 'error': return '错误';
      default: return '未知';
    }
  };

  const getStatusIcon = (status: string) => {
    switch (status) {
      case 'ready': return <CheckCircleOutlined />;
      case 'printing': return <PlayCircleOutlined />;
      case 'offline': return <StopOutlined />;
      case 'error': return <ExclamationCircleOutlined />;
      default: return <StopOutlined />;
    }
  };


  const columns = [
    {
      title: '打印机名称',
      dataIndex: 'name',
      key: 'name',
      render: (text: string) => <strong>{text}</strong>,
      width: 200,
    },
    {
      title: '型号',
      dataIndex: 'model',
      key: 'model',
      width: 200,
    },
    {
      title: '所属边缘节点',
      dataIndex: 'edge_node_name',
      key: 'edge_node_name',
      width: 180,
      render: (text: string) => text || '未知',
    },
    {
      title: '位置',
      dataIndex: 'location',
      key: 'location',
      width: 150,
      render: (text: string) => text || '-',
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
      title: '操作',
      key: 'action',
      width: 100,
      render: (_, record: PrinterStatus) => (
        <Space size="small">
          <Popconfirm
            title="确认删除"
            description={`确定要删除打印机 "${record.name}" 吗？`}
            onConfirm={() => handleDeletePrinter(record.id, record.name)}
            okText="确认"
            cancelText="取消"
          >
            <Button 
              type="text" 
              danger 
              size="small"
              icon={<DeleteOutlined />}
            >
              删除
            </Button>
          </Popconfirm>
        </Space>
      ),
    },
  ];

  return (
    <div>
      <div style={{ marginBottom: 16, display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
        <h2 style={{ margin: 0 }}>打印机管理</h2>
        <Space>
          <a onClick={() => window.location.reload()}>刷新</a>
        </Space>
      </div>

      {/* 筛选器 */}
      <div style={{ marginBottom: 16 }}>
        <Space>
          <span>边缘节点：</span>
          <Select 
            value={selectedEdgeNode} 
            onChange={setSelectedEdgeNode}
            style={{ width: 200 }}
            placeholder="选择边缘节点"
          >
            <Select.Option value="">全部</Select.Option>
            {edgeNodes.map(node => (
              <Select.Option key={node.id} value={node.id}>{node.name}</Select.Option>
            ))}
          </Select>
        </Space>
      </div>

      <Card>
        <Table
          columns={columns}
          dataSource={filteredPrinters}
          loading={loading}
          pagination={{
            current: currentPage,
            pageSize: pageSize,
            total: filteredPrinters.length,
            showSizeChanger: true,
            showQuickJumper: true,
            showTotal: (total) => `共 ${total} 台打印机`,
            onChange: (page, size) => {
              setCurrentPage(page);
              setPageSize(size || 10);
            },
            onShowSizeChange: (current, size) => {
              setCurrentPage(1);
              setPageSize(size);
            },
            pageSizeOptions: ['10', '20', '50', '100'],
          }}
          scroll={{ x: 900 }}
          locale={{
            emptyText: '暂无打印机数据'
          }}
        />
      </Card>

      {/* 统计信息 */}
      <Row gutter={16} style={{ marginTop: 16 }}>
        <Col span={6}>
          <Card>
            <Statistic
              title="总打印机数"
              value={filteredPrinters.length}
              prefix={<PrinterOutlined />}
            />
          </Card>
        </Col>
        <Col span={6}>
          <Card>
            <Statistic
              title="就绪打印机"
              value={filteredPrinters.filter(printer => printer.status === 'ready').length}
              prefix={<CheckCircleOutlined style={{ color: '#52c41a' }} />}
              valueStyle={{ color: '#52c41a' }}
            />
          </Card>
        </Col>
        <Col span={6}>
          <Card>
            <Statistic
              title="正在打印"
              value={filteredPrinters.filter(printer => printer.status === 'printing').length}
              prefix={<PlayCircleOutlined style={{ color: '#1890ff' }} />}
              valueStyle={{ color: '#1890ff' }}
            />
          </Card>
        </Col>
        <Col span={6}>
          <Card>
            <Statistic
              title="异常打印机"
              value={filteredPrinters.filter(printer => printer.status === 'error').length}
              prefix={<ExclamationCircleOutlined style={{ color: '#ff4d4f' }} />}
              valueStyle={{ color: '#ff4d4f' }}
            />
          </Card>
        </Col>
      </Row>
    </div>
  );
};

export default Printers;
