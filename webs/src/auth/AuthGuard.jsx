import PropTypes from 'prop-types';
import { Navigate, useLocation } from 'react-router-dom';

// project imports
import { useAuth } from 'contexts/AuthContext';
import Loader from 'ui-component/Loader';

// ==============================|| AUTH GUARD ||============================== //

/**
 * 路由守卫组件
 * 确保用户已登录才能访问受保护的页面
 */
export default function AuthGuard({ children }) {
  const { isAuthenticated, isInitialized } = useAuth();
  const location = useLocation();

  // 等待初始化完成
  if (!isInitialized) {
    return <Loader />;
  }

  // 未登录则跳转到登录页
  if (!isAuthenticated) {
    return <Navigate to="/login" state={{ from: location }} replace />;
  }

  return children;
}

AuthGuard.propTypes = {
  children: PropTypes.node.isRequired
};
