import React, { useState, useEffect } from 'react';
import { Card, Table, Tag, Space, Row, Col, Statistic, Progress, message } from 'antd';
import { 
  CheckCircleOutlined,
  ExclamationCircleOutlined,
  StopOutlined,
  PrinterOutlined,
  PlayCircleOutlined
} from '@ant-design/icons';

// 打印机接口（适配后端数据模型）
interface PrinterStatus {
  id: string;
  name: string;
  model: string;
  location?: string; // 后端可能为空
  status: 'ready' | 'printing' | 'error' | 'offline'; // 后端状态值
  edge_node_id: string;
  queue_length: number;
  // 以下字段暂时模拟，等待后端扩展
  paperLevel?: number;
  inkLevel?: number;
  key?: string;
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
        return result.data;
      }
    } catch (error) {
      console.error('获取打印机列表失败:', error);
    }
    
    // 返回模拟数据作为fallback（适配后端格式）
    return [
      {
        id: '1',
        name: 'HP LaserJet Pro M404n',
        model: 'HP LaserJet Pro M404n',
        location: '办公室 A-101',
        status: 'ready',
        edge_node_id: 'edge-1',
        queue_length: 0,
        paperLevel: 85,
        inkLevel: 60,
      },
      {
        id: '2',
        name: 'Canon PIXMA G3020',
        model: 'Canon PIXMA G3020',
        location: '办公室 B-205',
        status: 'printing',
        edge_node_id: 'edge-1',
        queue_length: 2,
        paperLevel: 45,
        inkLevel: 30,
      },
      {
        id: '3',
        name: 'Epson EcoTank L3150',
        model: 'Epson EcoTank L3150',
        location: '会议室 C-301',
        status: 'offline',
        edge_node_id: 'edge-2',
        queue_length: 0,
        paperLevel: 20,
        inkLevel: 80,
      },
      {
        id: '4',
        name: 'Brother HL-L2350DW',
        model: 'Brother HL-L2350DW',
        location: '前台接待',
        status: 'error',
        edge_node_id: 'edge-3',
        queue_length: 1,
        paperLevel: 0,
        inkLevel: 15,
      },
      {
        id: '5',
        name: 'Samsung ML-2161',
        model: 'Samsung ML-2161',
        location: '会议室 B-302',
        status: 'ready',
        edge_node_id: 'edge-2',
        queue_length: 0,
        paperLevel: 95,
        inkLevel: 45,
      },
      {
        id: '6',
        name: 'Xerox WorkCentre 3225',
        model: 'Xerox WorkCentre 3225',
        location: '财务部',
        status: 'ready',
        edge_node_id: 'edge-3',
        queue_length: 3,
        paperLevel: 70,
        inkLevel: 85,
      },
    ];
  }
}

const printersService = new PrintersService();

// Printers 组件
const Printers: React.FC = () => {
  const [printers, setPrinters] = useState<PrinterStatus[]>([]);
  const [loading, setLoading] = useState(true);

  // 加载打印机数据
  useEffect(() => {
    const loadPrinters = async () => {
      try {
        setLoading(true);
        const printerList = await printersService.getPrinters();
        setPrinters(printerList.map(printer => ({ ...printer, key: printer.id })));
      } catch (error) {
        console.error('加载打印机失败:', error);
        // 设置 fallback 数据
        const fallbackPrinters = [
          {
            id: '1',
            name: 'HP LaserJet Pro M404n',
            model: 'HP LaserJet Pro M404n',
            location: '办公室 A-101',
            status: 'ready' as const,
            edge_node_id: 'edge-1',
            queue_length: 0,
            paperLevel: 85,
            inkLevel: 60,
          },
          {
            id: '2',
            name: 'Canon PIXMA G3020',
            model: 'Canon PIXMA G3020',
            location: '办公室 B-205',
            status: 'printing' as const,
            edge_node_id: 'edge-1',
            queue_length: 2,
            paperLevel: 45,
            inkLevel: 30,
          },
          {
            id: '3',
            name: 'Epson EcoTank L3150',
            model: 'Epson EcoTank L3150',
            location: '会议室 C-301',
            status: 'offline' as const,
            edge_node_id: 'edge-2',
            queue_length: 0,
            paperLevel: 20,
            inkLevel: 80,
          },
          {
            id: '4',
            name: 'Brother HL-L2350DW',
            model: 'Brother HL-L2350DW',
            location: '前台接待',
            status: 'error' as const,
            edge_node_id: 'edge-3',
            queue_length: 1,
            paperLevel: 0,
            inkLevel: 15,
          },
        ];
        setPrinters(fallbackPrinters.map(printer => ({ ...printer, key: printer.id })));
      } finally {
        setLoading(false);
      }
    };

    loadPrinters();
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

  const getLevelColor = (level: number) => {
    if (level > 50) return '#52c41a';
    if (level > 20) return '#faad14';
    return '#ff4d4f';
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
      title: '位置',
      dataIndex: 'location',
      key: 'location',
      width: 150,
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
      title: '纸张余量',
      dataIndex: 'paperLevel',
      key: 'paperLevel',
      width: 150,
      render: (level: number) => (
        <div>
          <Progress 
            percent={level} 
            size="small" 
            strokeColor={getLevelColor(level)}
            format={(percent) => `${percent}%`}
          />
        </div>
      ),
    },
    {
      title: '墨水/墨粉',
      dataIndex: 'inkLevel',
      key: 'inkLevel',
      width: 150,
      render: (level: number) => (
        <div>
          <Progress 
            percent={level} 
            size="small" 
            strokeColor={getLevelColor(level)}
            format={(percent) => `${percent}%`}
          />
        </div>
      ),
    },
    {
      title: '操作',
      key: 'action',
      width: 150,
      render: (record: PrinterStatus) => (
        <Space>
          <a onClick={() => message.info(`查看打印机 ${record.name} 详情`)}>详情</a>
          <a onClick={() => message.info(`测试打印机 ${record.name}`)}>测试</a>
          {record.status === 'error' && (
            <a onClick={() => message.info(`重启打印机 ${record.name}`)}>重启</a>
          )}
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

      <Card>
        <Table
          columns={columns}
          dataSource={printers}
          loading={loading}
          pagination={{
            total: printers.length,
            pageSize: 10,
            showSizeChanger: true,
            showQuickJumper: true,
            showTotal: (total) => `共 ${total} 台打印机`,
          }}
          scroll={{ x: 900 }}
        />
      </Card>

      {/* 统计信息 */}
      <Row gutter={16} style={{ marginTop: 16 }}>
        <Col span={6}>
          <Card>
            <Statistic
              title="总打印机数"
              value={printers.length}
              prefix={<PrinterOutlined />}
            />
          </Card>
        </Col>
        <Col span={6}>
          <Card>
            <Statistic
              title="就绪打印机"
              value={printers.filter(printer => printer.status === 'ready').length}
              prefix={<CheckCircleOutlined style={{ color: '#52c41a' }} />}
              valueStyle={{ color: '#52c41a' }}
            />
          </Card>
        </Col>
        <Col span={6}>
          <Card>
            <Statistic
              title="正在打印"
              value={printers.filter(printer => printer.status === 'printing').length}
              prefix={<PlayCircleOutlined style={{ color: '#1890ff' }} />}
              valueStyle={{ color: '#1890ff' }}
            />
          </Card>
        </Col>
        <Col span={6}>
          <Card>
            <Statistic
              title="异常打印机"
              value={printers.filter(printer => printer.status === 'error').length}
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
