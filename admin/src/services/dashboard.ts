// Dashboard 数据接口定义
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
  paperLevel?: number;
  inkLevel?: number;
  key?: string;
}

export interface PrintJob {
  id: string;
  fileName: string;
  user: string;
  printer: string;
  status: 'pending' | 'printing' | 'completed' | 'failed';
  createdAt: string;
  pages: number;
  key?: string;
}

export interface EdgeNodeStatus {
  id: string;
  name: string;
  location: string;
  status: 'online' | 'offline';
  lastSeen: string;
  version: string;
  printerCount: number;
  key?: string;
}

import apiService from './api';

// Dashboard 服务类
class DashboardService {
  // 获取统计数据
  async getStats(): Promise<DashboardStats> {
    try {
      const response = await apiService.get<DashboardStats>('/admin/dashboard/stats');
      return response.data;
    } catch (error) {
      console.error('获取统计数据失败:', error);
      // API调用失败时返回空数据
      return {
        totalPrinters: 0,
        onlinePrinters: 0,
        totalEdgeNodes: 0,
        onlineEdgeNodes: 0,
        totalPrintJobs: 0,
        completedJobs: 0,
        totalUsers: 0,
        activeUsers: 0,
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
      return [];
    }
  }

  // 获取打印任务列表
  async getPrintJobs(page = 1, pageSize = 10): Promise<{ jobs: PrintJob[]; total: number }> {
    try {
      const response = await apiService.get<{ data: PrintJob[]; total: number }>(`/admin/print-jobs?page=${page}&pageSize=${pageSize}`);
      return {
        jobs: response.data.data.map(job => ({
          ...job,
          key: job.id, // 为 Ant Design Table 添加 key
        })),
        total: response.data.total,
      };
    } catch (error) {
      console.error('获取打印任务列表失败:', error);
      return {
        jobs: [],
        total: 0
      };
    }
  }

  // 获取边缘节点列表  
  async getEdgeNodes(): Promise<EdgeNodeStatus[]> {
    try {
      const response = await apiService.get<EdgeNodeStatus[]>('/admin/edge-nodes');
      return response.data.map(node => ({
        ...node,
        key: node.id, // 为 Ant Design Table 添加 key
      }));
    } catch (error) {
      console.error('获取边缘节点列表失败:', error);
      return [];
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
      return {
        dates: [],
        completed: [],
        failed: []
      };
    }
  }
}

// 创建单例实例
const dashboardService = new DashboardService();

export default dashboardService;