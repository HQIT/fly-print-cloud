import React, { useState, useEffect } from 'react';
import { Card, Table, Tag, Space, Row, Col, Statistic, message, Modal, Form, Input, Button } from 'antd';
import { 
  CheckCircleOutlined,
  ExclamationCircleOutlined,
  StopOutlined,
  PrinterOutlined,
  CloudServerOutlined,
  EditOutlined
} from '@ant-design/icons';

// 边缘节点接口（适配后端数据模型）
interface EdgeNode {
  id: string;
  name: string;
  location: string;
  status: 'online' | 'offline' | 'error';
  enabled: boolean;
  last_heartbeat: string;
  version: string;
  printer_count: number;  // 后端返回的打印机数量字段
  key?: string;
}

// Edge Nodes 服务类
class EdgeNodesService {
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
        console.log('🔄 [DEBUG] API响应数据:', result);
        
        // 适配后端数据格式：result.data.items
        return result?.data?.items || [];
      } else {
        console.error('💥 [DEBUG] API响应状态:', response.status, response.statusText);
      }
    } catch (error) {
      console.error('💥 [DEBUG] 网络请求异常:', error);
    }
    
    console.log('🔄 [DEBUG] API调用失败，返回空数据');
    return [];
  }

  async updateEdgeNode(id: string, name: string): Promise<boolean> {
    try {
      const token = await this.getToken();
      const response = await fetch(`/api/v1/admin/edge-nodes/${id}`, {
        method: 'PUT',
        headers: {
          'Content-Type': 'application/json',
          ...(token && { 'Authorization': `Bearer ${token}` }),
        },
        body: JSON.stringify({ name: name.trim() }),
      });
      
      return response.ok;
    } catch (error) {
      console.error('更新Edge Node失败:', error);
      return false;
    }
  }

  async updateEdgeNodeEnabled(id: string, enabled: boolean): Promise<boolean> {
    try {
      const token = await this.getToken();
      // 先获取当前的Edge Node信息
      const nodes = await this.getEdgeNodes();
      const currentNode = nodes.find(node => node.id === id);
      if (!currentNode) {
        console.error('Edge Node not found:', id);
        return false;
      }

      const response = await fetch(`/api/v1/admin/edge-nodes/${id}`, {
        method: 'PUT',
        headers: {
          'Content-Type': 'application/json',
          ...(token && { 'Authorization': `Bearer ${token}` }),
        },
        body: JSON.stringify({ name: currentNode.name, enabled }),
      });
      
      return response.ok;
    } catch (error) {
      console.error('更新Edge Node启用状态失败:', error);
      return false;
    }
  }
}

const edgeNodesService = new EdgeNodesService();

// Edge Nodes 组件
const EdgeNodes: React.FC = () => {
  const [edgeNodes, setEdgeNodes] = useState<EdgeNode[]>([]);
  const [loading, setLoading] = useState(true);
  
  // 编辑相关状态
  const [editModalVisible, setEditModalVisible] = useState(false);
  const [editingNode, setEditingNode] = useState<EdgeNode | null>(null);
  const [form] = Form.useForm();

  // 加载边缘节点数据
  const loadEdgeNodes = async () => {
    try {
      setLoading(true);
      const nodes = await edgeNodesService.getEdgeNodes();
      setEdgeNodes(nodes.map(node => ({ ...node, key: node.id })));
    } catch (error) {
      console.error('加载边缘节点失败:', error);
      setEdgeNodes([]);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    loadEdgeNodes();
  }, []);

  // 编辑Edge Node名称
  const handleEditNode = (node: EdgeNode) => {
    setEditingNode(node);
    form.setFieldsValue({ name: node.name });
    setEditModalVisible(true);
  };

  // 提交名称修改
  const handleEditSubmit = async (values: { name: string }) => {
    if (!editingNode) return;

    try {
      const success = await edgeNodesService.updateEdgeNode(editingNode.id, values.name);
      if (success) {
        message.success('Edge Node名称修改成功');
        setEditModalVisible(false);
        setEditingNode(null);
        form.resetFields();
        loadEdgeNodes(); // 重新加载数据
      } else {
        message.error('修改失败，请稍后重试');
      }
    } catch (error) {
      console.error('修改Edge Node名称失败:', error);
      message.error('修改失败，请稍后重试');
    }
  };

  // 切换启用/禁用状态
  const handleToggleEnabled = async (node: EdgeNode) => {
    try {
      const newEnabled = !node.enabled;
      const success = await edgeNodesService.updateEdgeNodeEnabled(node.id, newEnabled);
      if (success) {
        message.success(`Edge Node已${newEnabled ? '启用' : '禁用'}`);
        loadEdgeNodes(); // 重新加载数据
      } else {
        message.error('操作失败，请稍后重试');
      }
    } catch (error) {
      console.error('切换Edge Node状态失败:', error);
      message.error('操作失败，请稍后重试');
    }
  };

  // 状态图标映射
  const getStatusIcon = (status: string) => {
    switch (status) {
      case 'online':
        return <CheckCircleOutlined style={{ color: '#52c41a' }} />;
      case 'offline':
        return <StopOutlined style={{ color: '#8c8c8c' }} />;
      case 'error':
        return <ExclamationCircleOutlined style={{ color: '#ff4d4f' }} />;
      default:
        return <StopOutlined style={{ color: '#8c8c8c' }} />;
    }
  };

  // 状态标签映射
  const getStatusTag = (status: string) => {
    switch (status) {
      case 'online':
        return <Tag color="success">在线</Tag>;
      case 'offline':
        return <Tag color="default">离线</Tag>;
      case 'error':
        return <Tag color="error">错误</Tag>;
      default:
        return <Tag color="default">未知</Tag>;
    }
  };

  // 表格列定义
  const columns = [
    {
      title: '节点名称',
      dataIndex: 'name',
      key: 'name',
      render: (text: string) => (
        <Space>
          <CloudServerOutlined />
          {text}
        </Space>
      ),
    },
    {
      title: '位置',
      dataIndex: 'location',
      key: 'location',
      render: (text: string) => text || '-',
    },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      render: (status: string) => (
        <Space>
          {getStatusIcon(status)}
          {getStatusTag(status)}
        </Space>
      ),
    },
    {
      title: '最后心跳',
      dataIndex: 'last_heartbeat',
      key: 'last_heartbeat',
      render: (time: string) => {
        if (!time) return '-';
        const date = new Date(time);
        return date.toLocaleString('zh-CN');
      },
    },
    {
      title: '版本',
      dataIndex: 'version',
      key: 'version',
      render: (text: string) => text || '-',
    },
    {
      title: '打印机数量',
      dataIndex: 'printer_count',
      key: 'printer_count',
      render: (count: number) => (
        <Space>
          <PrinterOutlined />
          {count || 0}
        </Space>
      ),
    },
    {
      title: '操作',
      key: 'action',
      width: 120,
      render: (_, record: EdgeNode) => (
        <Space size="small">
          <Button 
            type="text" 
            size="small"
            icon={<EditOutlined />}
            onClick={() => handleEditNode(record)}
          >
            编辑名称
          </Button>
          <Button 
            type="text" 
            size="small"
            onClick={() => handleToggleEnabled(record)}
            style={{ 
              color: record.enabled ? '#ff4d4f' : '#52c41a' 
            }}
          >
            {record.enabled ? '禁用' : '启用'}
          </Button>
        </Space>
      ),
    },
  ];

  // 计算统计数据
  const onlineNodes = edgeNodes.filter(node => node.status === 'online').length;
  const offlineNodes = edgeNodes.filter(node => node.status === 'offline').length;
  const errorNodes = edgeNodes.filter(node => node.status === 'error').length;
  const totalPrinters = edgeNodes.reduce((sum, node) => sum + (node.printer_count || 0), 0);

  return (
    <div style={{ padding: '24px' }}>
      <h2>边缘节点管理</h2>
      
      {/* 统计卡片 */}
      <Row gutter={[16, 16]} style={{ marginBottom: '24px' }}>
        <Col xs={12} sm={6}>
          <Card>
            <Statistic
              title="总节点数"
              value={edgeNodes.length}
              prefix={<CloudServerOutlined />}
              valueStyle={{ color: '#1890ff' }}
            />
          </Card>
        </Col>
        <Col xs={12} sm={6}>
          <Card>
            <Statistic
              title="在线节点"
              value={onlineNodes}
              prefix={<CheckCircleOutlined />}
              valueStyle={{ color: '#52c41a' }}
            />
          </Card>
        </Col>
        <Col xs={12} sm={6}>
          <Card>
            <Statistic
              title="离线节点"
              value={offlineNodes}
              prefix={<StopOutlined />}
              valueStyle={{ color: '#8c8c8c' }}
            />
          </Card>
        </Col>
        <Col xs={12} sm={6}>
          <Card>
            <Statistic
              title="总打印机数"
              value={totalPrinters}
              prefix={<PrinterOutlined />}
              valueStyle={{ color: '#722ed1' }}
            />
          </Card>
        </Col>
      </Row>

      {/* 边缘节点列表 */}
      <Card title="边缘节点列表">
        <Table
          columns={columns}
          dataSource={edgeNodes}
          loading={loading}
          pagination={{
            total: edgeNodes.length,
            pageSize: 10,
            showSizeChanger: true,
            showQuickJumper: true,
            showTotal: (total, range) =>
              `第 ${range[0]}-${range[1]} 条，共 ${total} 条`,
          }}
          size="middle"
        />
      </Card>

      {/* 编辑Edge Node名称模态框 */}
      <Modal
        title="编辑Edge Node名称"
        open={editModalVisible}
        onCancel={() => {
          setEditModalVisible(false);
          setEditingNode(null);
          form.resetFields();
        }}
        footer={null}
        width={500}
      >
        <Form
          form={form}
          layout="vertical"
          onFinish={handleEditSubmit}
        >
          <Form.Item
            name="name"
            label="节点名称"
            rules={[
              { required: true, message: '请输入节点名称' },
              { max: 100, message: '名称不能超过100个字符' }
            ]}
          >
            <Input placeholder="输入节点名称" />
          </Form.Item>
          
          {editingNode && (
            <div style={{ marginBottom: 16, padding: 12, backgroundColor: '#f5f5f5', borderRadius: 6 }}>
              <div><strong>节点ID：</strong>{editingNode.id}</div>
              <div><strong>当前状态：</strong>{editingNode.status}</div>
            </div>
          )}

          <Form.Item style={{ marginBottom: 0, textAlign: 'right' }}>
            <Space>
              <Button onClick={() => {
                setEditModalVisible(false);
                setEditingNode(null);
                form.resetFields();
              }}>
                取消
              </Button>
              <Button type="primary" htmlType="submit">
                保存
              </Button>
            </Space>
          </Form.Item>
        </Form>
      </Modal>
    </div>
  );
};

export default EdgeNodes;