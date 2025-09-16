import React, { useState, useEffect } from 'react';
import { BrowserRouter as Router, Routes, Route, Navigate, useNavigate, useLocation } from 'react-router-dom';
import { Layout, Menu, Avatar, Dropdown, Space, Spin, message, Alert } from 'antd';
import { 
  DashboardOutlined, 
  PrinterOutlined, 
  CloudServerOutlined, 
  FileTextOutlined,
  UserOutlined,
  LogoutOutlined,
  SettingOutlined
} from '@ant-design/icons';
import type { MenuProps } from 'antd';

// 导入页面组件
import Dashboard from './components/pages/Dashboard.tsx';
import EdgeNodes from './components/pages/EdgeNodes.tsx';
import Printers from './components/pages/Printers.tsx';
import PrintJobs from './components/pages/PrintJobs.tsx';
import Users from './components/pages/Users.tsx';
import Settings from './components/pages/Settings.tsx';

const { Header, Sider, Content } = Layout;

interface User {
  id: string;
  username: string;
  email: string;
  role: string;
  status: string;
}

// 主应用组件
const AppContent: React.FC = () => {
  const navigate = useNavigate();
  const location = useLocation();
  
  const [collapsed, setCollapsed] = useState(false);
  const [user, setUser] = useState<User | null>(null);
  const [loading, setLoading] = useState(true);

  // 获取当前用户信息
  useEffect(() => {
    const getCurrentUser = async () => {
      try {
        const response = await fetch('/auth/me');
        const result = await response.json();
        
        if (result.code === 200) {
          setUser({
            id: result.data.user_id || '1',
            username: result.data.preferred_username || result.data.username || 'n/a',
            email: result.data.email || 'admin@example.com',
            role: 'admin',
            status: 'active'
          });
        } else {
          // 如果获取用户信息失败，重定向到登录页面
          window.location.href = '/auth/login';
        }
      } catch (error) {
        console.error('获取用户信息失败:', error);
        // 网络错误时也重定向到登录页面
        window.location.href = '/auth/login';
      } finally {
        setLoading(false);
      }
    };

    getCurrentUser();
  }, []);

  // 处理登出
  const handleLogout = async () => {
    try {
      await fetch('/auth/logout', { method: 'POST' });
    } catch (error) {
      console.error('登出失败:', error);
    } finally {
      window.location.href = '/auth/login';
    }
  };

  // 菜单项配置
  const menuItems: MenuProps['items'] = [
    {
      key: '/',
      icon: <DashboardOutlined />,
      label: 'Dashboard',
    },
    {
      key: '/edge-nodes',
      icon: <CloudServerOutlined />,
      label: 'Edge Nodes',
    },
    {
      key: '/printers',
      icon: <PrinterOutlined />,
      label: 'Printers',
    },
    {
      key: '/print-jobs',
      icon: <FileTextOutlined />,
      label: 'Print Jobs',
    },
    {
      key: '/users',
      icon: <UserOutlined />,
      label: 'Users',
    },
    {
      key: '/settings',
      icon: <SettingOutlined />,
      label: 'Settings',
    },
  ];

  // 用户下拉菜单
  const userMenuItems: MenuProps['items'] = [
    {
      key: 'profile',
      icon: <UserOutlined />,
      label: '个人资料',
    },
    {
      key: 'settings',
      icon: <SettingOutlined />,
      label: '设置',
    },
    {
      type: 'divider',
    },
    {
      key: 'logout',
      icon: <LogoutOutlined />,
      label: '退出登录',
      onClick: handleLogout,
    },
  ];

  const handleMenuClick = (e: any) => {
    navigate(e.key);
  };

  if (loading) {
    return (
      <div style={{ 
        display: 'flex', 
        justifyContent: 'center', 
        alignItems: 'center', 
        height: '100vh' 
      }}>
        <Spin size="large" tip="加载中..." />
      </div>
    );
  }

  return (
    <Layout style={{ minHeight: '100vh' }}>
      <Sider
        collapsible
        collapsed={collapsed}
        onCollapse={setCollapsed}
        style={{
          position: 'fixed',
          height: '100vh',
          left: 0,
          top: 0,
          bottom: 0,
        }}
      >
        <div style={{
          height: 32,
          margin: 16,
          background: 'rgba(255, 255, 255, 0.3)',
          borderRadius: 6,
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center',
          color: 'white',
          fontWeight: 'bold',
        }}>
          {collapsed ? 'FP' : 'FlyPrint'}
        </div>
        <Menu
          theme="dark"
          selectedKeys={[location.pathname]}
          mode="inline"
          items={menuItems}
          onClick={handleMenuClick}
        />
      </Sider>

      <Layout style={{ marginLeft: collapsed ? 80 : 200, transition: 'margin-left 0.2s' }}>
        <Header style={{
          background: '#fff',
          padding: '0 24px',
          display: 'flex',
          justifyContent: 'space-between',
          alignItems: 'center',
          boxShadow: '0 1px 4px rgba(0,21,41,.08)',
          position: 'sticky',
          top: 0,
          zIndex: 1,
        }}>
          <div />
          <Space>
            <span>欢迎, {user?.username}</span>
            <Dropdown menu={{ items: userMenuItems }} placement="bottomRight">
              <Avatar style={{ cursor: 'pointer' }}>
                {user?.username?.charAt(0).toUpperCase()}
              </Avatar>
            </Dropdown>
          </Space>
        </Header>

        <Content style={{
          margin: '24px',
          minHeight: 'calc(100vh - 112px)',
        }}>
          <div style={{
            background: '#fff',
            padding: 24,
            borderRadius: 8,
            boxShadow: '0 1px 3px rgba(0, 0, 0, 0.1)',
            minHeight: 'calc(100vh - 160px)',
          }}>
            <Routes>
              <Route path="/" element={<Dashboard />} />
              <Route path="/edge-nodes" element={<EdgeNodes />} />
              <Route path="/printers" element={<Printers />} />
              <Route path="/print-jobs" element={<PrintJobs />} />
              <Route path="/users" element={<Users />} />
              <Route path="/settings" element={<Settings />} />
              <Route path="*" element={<Navigate to="/" replace />} />
            </Routes>
          </div>
        </Content>
      </Layout>
    </Layout>
  );
};

// App 根组件
const App: React.FC = () => {
  return (
    <Router>
      <AppContent />
    </Router>
  );
};

export default App;