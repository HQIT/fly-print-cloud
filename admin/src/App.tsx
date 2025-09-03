import React, { useState, useEffect } from 'react';

interface User {
  id: string;
  username: string;
  email: string;
  role: string;
  status: string;
}

function App() {
  const [user, setUser] = useState<User | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    // 1. 先获取认证信息和 access_token
    fetch('/auth/me')
      .then(res => res.json())
      .then(authData => {
        if (authData.code !== 200 || !authData.data.access_token) {
          setLoading(false);
          return;
        }

        const accessToken = authData.data.access_token;

        // 2. 使用 access_token 获取业务用户信息
        return fetch('/api/v1/admin/profile', {
          headers: {
            'Authorization': `Bearer ${accessToken}`
          }
        });
      })
      .then(res => res?.json())
      .then(data => {
        if (data?.code === 200) {
          setUser(data.data);
        }
        setLoading(false);
      })
      .catch(err => {
        console.error('获取用户信息失败:', err);
        setLoading(false);
      });
  }, []);

  const handleLogout = () => {
    window.location.href = '/auth/logout';
  };

  if (loading) {
    return <div>加载中...</div>;
  }

  return (
    <div className="App">
      <header style={{ 
        display: 'flex', 
        justifyContent: 'space-between', 
        alignItems: 'center',
        padding: '1rem 2rem',
        borderBottom: '1px solid #ddd',
        backgroundColor: '#f8f9fa'
      }}>
        <div>
          <h1>Fly Print Cloud - Admin Console</h1>
          <p>云打印管理系统</p>
        </div>
        {user && (
          <div style={{ textAlign: 'right' }}>
            <div>欢迎, {user.username}</div>
            <div style={{ fontSize: '0.9em', color: '#666' }}>{user.email}</div>
            <button 
              onClick={handleLogout}
              style={{ 
                marginTop: '0.5rem',
                padding: '0.5rem 1rem',
                backgroundColor: '#dc3545',
                color: 'white',
                border: 'none',
                borderRadius: '4px',
                cursor: 'pointer'
              }}
            >
              登出
            </button>
          </div>
        )}
      </header>
      <main style={{ padding: '2rem' }}>
        {user ? (
          <div>
            <h2>管理功能</h2>
            <p>当前角色: {user.role}</p>
            <p>账户状态: {user.status}</p>
          </div>
        ) : (
          <div>未登录或用户信息加载失败</div>
        )}
      </main>
    </div>
  );
}

export default App;
