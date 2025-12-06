import { Outlet } from 'react-router-dom';
import { AuthProvider } from 'contexts/AuthContext';

// ==============================|| ROUTER WRAPPER WITH AUTH ||============================== //

/**
 * 路由包装组件
 * 为路由提供 AuthProvider 上下文
 */
export default function RouterWrapper() {
  return (
    <AuthProvider>
      <Outlet />
    </AuthProvider>
  );
}
