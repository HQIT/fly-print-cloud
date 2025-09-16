import apiService, { ApiResponse } from './api';

// Dashboard 数据接口
export interface DashboardStats {
  totalPrinters: number;
  onlinePrinters: number;
  totalEdgeNodes: number;
  onlineEdgeNodes: number;
  totalPrintJobs: number;
  completedJobs: number;
  totalUsers: number;
  activeUsers: number;
}

export interface PrinterStatus {
  id: string;
  name: string;
  location: string;
  status: 'online' | 'offline' | 'printing' | 'error';
  paperLevel: number;
  inkLevel: number;
  lastSeen?: string;
  model?: string;
}

export interface PrintJob {
  id: string;
  fileName: string;
  user: string;
  printer: string;
  status: 'pending' | 'printing' | 'completed' | 'failed';
  createdAt: string;
  completedAt?: string;
  pages: number;
  copies?: number;
  priority?: 'low' | 'normal' | 'high';
}

export interface EdgeNode {
  id: string;
  name: string;
  location: string;
  status: 'online' | 'offline' | 'error';
  lastSeen: string;
  version: string;
  printerCount: number;
}

// Dashboard 服务类
class DashboardService {
  // 获取统计数据
  async getStats(): Promise<DashboardStats> {
    try {
      const response = await apiService.get<DashboardStats>('/admin/dashboard/stats');
      return response.data;
    } catch (error) {
      console.error('获取统计数据失败:', error);
      // 返回模拟数据作为fallback
      return {
        totalPrinters: 12,
        onlinePrinters: 9,
        totalEdgeNodes: 5,
        onlineEdgeNodes: 4,
        totalPrintJobs: 156,
        completedJobs: 142,
        totalUsers: 48,
        activeUsers: 23,
      };
    }
  }

  // 获取打印机列表
  async getPrinters(): Promise<PrinterStatus[]> {
    try {
      const response = await apiService.get<PrinterStatus[]>('/admin/printers');
      return response.data.map(printer => ({
        ...printer,
        key: printer.id, // 为 Ant Design Table 添加 key
      }));
    } catch (error) {
      console.error('获取打印机列表失败:', error);
      // 返回模拟数据作为fallback
      return [
        {
          id: '1',
          name: 'HP LaserJet Pro M404n',
          location: '办公室 A-101',
          status: 'online',
          paperLevel: 85,
          inkLevel: 60,
        },
        {
          id: '2',
          name: 'Canon PIXMA G3020',
          location: '办公室 B-205',
          status: 'printing',
          paperLevel: 45,
          inkLevel: 30,
        },
        {
          id: '3',
          name: 'Epson EcoTank L3150',
          location: '会议室 C-301',
          status: 'offline',
          paperLevel: 20,
          inkLevel: 80,
        },
        {
          id: '4',
          name: 'Brother HL-L2350DW',
          location: '前台接待',
          status: 'error',
          paperLevel: 0,
          inkLevel: 15,
        },
      ];
    }
  }

  // 获取打印任务列表
  async getPrintJobs(page = 1, pageSize = 10): Promise<{ jobs: PrintJob[]; total: number }> {
    try {
      const response = await apiService.get<{ jobs: PrintJob[]; total: number }>(
        `/admin/print-jobs?page=${page}&pageSize=${pageSize}`
      );
      return {
        jobs: response.data.jobs.map(job => ({
          ...job,
          key: job.id, // 为 Ant Design Table 添加 key
        })),
        total: response.data.total,
      };
    } catch (error) {
      console.error('获取打印任务列表失败:', error);
      // 返回模拟数据作为fallback
      return {
        jobs: [
          {
            id: 'JOB-2024-001',
            fileName: '合同文件.pdf',
            user: '张三',
            printer: 'HP LaserJet Pro M404n',
            status: 'completed',
            createdAt: '2024-01-15 10:30',
            pages: 5,
          },
          {
            id: 'JOB-2024-002',
            fileName: '报告.docx',
            user: '李四',
            printer: 'Canon PIXMA G3020',
            status: 'printing',
            createdAt: '2024-01-15 11:15',
            pages: 12,
          },
          {
            id: 'JOB-2024-003',
            fileName: '图表分析.xlsx',
            user: '王五',
            printer: 'Epson EcoTank L3150',
            status: 'pending',
            createdAt: '2024-01-15 11:45',
            pages: 3,
          },
          {
            id: 'JOB-2024-004',
            fileName: '项目计划.pptx',
            user: '赵六',
            printer: 'Brother HL-L2350DW',
            status: 'failed',
            createdAt: '2024-01-15 12:00',
            pages: 20,
          },
        ],
        total: 4,
      };
    }
  }

  // 获取边缘节点列表
  async getEdgeNodes(): Promise<EdgeNode[]> {
    try {
      const response = await apiService.get<EdgeNode[]>('/admin/edge-nodes');
      return response.data.map(node => ({
        ...node,
        key: node.id, // 为 Ant Design Table 添加 key
      }));
    } catch (error) {
      console.error('获取边缘节点列表失败:', error);
      // 返回模拟数据作为fallback
      return [
        {
          id: '1',
          name: 'EdgeNode-Office-A',
          location: '办公楼A',
          status: 'online',
          lastSeen: '2024-01-15 12:00',
          version: 'v1.2.3',
          printerCount: 3,
        },
        {
          id: '2',
          name: 'EdgeNode-Office-B',
          location: '办公楼B',
          status: 'online',
          lastSeen: '2024-01-15 11:58',
          version: 'v1.2.3',
          printerCount: 2,
        },
      ];
    }
  }

  // 获取打印任务趋势数据
  async getPrintJobTrends(days = 7): Promise<{ dates: string[]; completed: number[]; failed: number[] }> {
    try {
      const response = await apiService.get<{ dates: string[]; completed: number[]; failed: number[] }>(
        `/admin/dashboard/print-job-trends?days=${days}`
      );
      return response.data;
    } catch (error) {
      console.error('获取打印任务趋势失败:', error);
      // 返回模拟数据作为fallback
      return {
        dates: ['周一', '周二', '周三', '周四', '周五', '周六', '周日'],
        completed: [12, 18, 15, 22, 28, 16, 20],
        failed: [2, 1, 3, 2, 1, 4, 2],
      };
    }
  }
}

// 创建单例实例
const dashboardService = new DashboardService();

export default dashboardService;
