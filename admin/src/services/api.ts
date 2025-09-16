// API 基础服务
export interface ApiResponse<T = any> {
  code: number;
  message: string;
  data: T;
}

export interface ApiError {
  code: number;
  message: string;
  details?: any;
}

class ApiService {
  private baseURL = '/api/v1';
  private token: string | null = null;

  // 设置认证 token
  setToken(token: string) {
    this.token = token;
  }

  // 获取认证 token
  async getToken(): Promise<string | null> {
    if (this.token) {
      return this.token;
    }

    try {
      const response = await fetch('/auth/me');
      const result = await response.json();
      
      if (result.code === 200 && result.data.access_token) {
        this.token = result.data.access_token;
        return this.token;
      }
    } catch (error) {
      console.error('获取 token 失败:', error);
    }
    
    return null;
  }

  // 通用请求方法
  private async request<T = any>(
    endpoint: string,
    options: RequestInit = {}
  ): Promise<ApiResponse<T>> {
    const token = await this.getToken();
    
    const config: RequestInit = {
      headers: {
        'Content-Type': 'application/json',
        ...(token && { 'Authorization': `Bearer ${token}` }),
        ...options.headers,
      },
      ...options,
    };

    try {
      const response = await fetch(`${this.baseURL}${endpoint}`, config);
      const result = await response.json();

      if (!response.ok) {
        throw new ApiError({
          code: response.status,
          message: result.message || '请求失败',
          details: result,
        });
      }

      return result;
    } catch (error) {
      if (error instanceof ApiError) {
        throw error;
      }
      
      throw new ApiError({
        code: 500,
        message: error instanceof Error ? error.message : '网络错误',
      });
    }
  }

  // GET 请求
  async get<T = any>(endpoint: string): Promise<ApiResponse<T>> {
    return this.request<T>(endpoint, { method: 'GET' });
  }

  // POST 请求
  async post<T = any>(endpoint: string, data?: any): Promise<ApiResponse<T>> {
    return this.request<T>(endpoint, {
      method: 'POST',
      body: data ? JSON.stringify(data) : undefined,
    });
  }

  // PUT 请求
  async put<T = any>(endpoint: string, data?: any): Promise<ApiResponse<T>> {
    return this.request<T>(endpoint, {
      method: 'PUT',
      body: data ? JSON.stringify(data) : undefined,
    });
  }

  // DELETE 请求
  async delete<T = any>(endpoint: string): Promise<ApiResponse<T>> {
    return this.request<T>(endpoint, { method: 'DELETE' });
  }
}

// 创建单例实例
const apiService = new ApiService();

// 扩展 Error 类来处理 API 错误
class ApiError extends Error {
  code: number;
  details?: any;

  constructor({ code, message, details }: { code: number; message: string; details?: any }) {
    super(message);
    this.name = 'ApiError';
    this.code = code;
    this.details = details;
  }
}

export { apiService, ApiError };
export default apiService;
