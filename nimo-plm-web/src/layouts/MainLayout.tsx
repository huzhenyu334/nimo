import React, { useMemo } from 'react';
import { Outlet, useNavigate, useLocation } from 'react-router-dom';
import { ProLayout, ProLayoutProps } from '@ant-design/pro-components';
import {
  HomeOutlined,
  ProjectOutlined,
  ExperimentOutlined,
  SnippetsOutlined,
  LogoutOutlined,
  UserOutlined,
  FolderOutlined,
} from '@ant-design/icons';
import { Dropdown, Avatar, Space, Spin } from 'antd';
import { useAuth } from '@/contexts/AuthContext';
import { useQuery } from '@tanstack/react-query';
import { projectApi } from '@/api/projects';

const MainLayout: React.FC = () => {
  const navigate = useNavigate();
  const location = useLocation();
  const { user, logout, isLoading } = useAuth();

  // 获取项目列表
  const { data: projectData } = useQuery({
    queryKey: ['projects'],
    queryFn: () => projectApi.list(),
    enabled: !!user,
  });

  const menuItems = useMemo(() => {
    const projects = projectData?.items || [];
    const projectChildren = projects.map((p: any) => ({
      path: `/projects/${p.id}`,
      name: `${p.name}`,
    }));

    return [
      {
        path: '/dashboard',
        name: '工作台',
        icon: <HomeOutlined />,
      },
      {
        path: '/projects',
        name: '研发项目',
        icon: <ProjectOutlined />,
        children: projectChildren.length > 0 ? [
          { path: '/projects', name: '全部项目', icon: <FolderOutlined /> },
          ...projectChildren,
        ] : undefined,
      },
      {
        path: '/materials',
        name: '物料选型库',
        icon: <ExperimentOutlined />,
      },
      {
        path: '/templates',
        name: '研发流程',
        icon: <SnippetsOutlined />,
      },
    ];
  }, [projectData]);

  if (isLoading) {
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

  // Highlight current menu
  const menuPathname = location.pathname;

  const layoutProps: ProLayoutProps = {
    title: 'nimo PLM',
    logo: '/logo.svg',
    layout: 'mix',
    splitMenus: false,
    fixedHeader: true,
    fixSiderbar: true,
    contentWidth: 'Fluid',
    route: {
      path: '/',
      routes: menuItems,
    },
    location: {
      pathname: menuPathname,
    },
    menuItemRender: (item, dom) => (
      <div onClick={() => item.path && navigate(item.path)}>{dom}</div>
    ),
    avatarProps: {
      src: user?.avatar_url,
      size: 'small',
      title: user?.name,
      render: (_, dom) => (
        <Dropdown
          menu={{
            items: [
              {
                key: 'profile',
                icon: <UserOutlined />,
                label: '个人信息',
              },
              {
                type: 'divider',
              },
              {
                key: 'logout',
                icon: <LogoutOutlined />,
                label: '退出登录',
                onClick: () => {
                  logout();
                  navigate('/login');
                },
              },
            ],
          }}
        >
          {dom}
        </Dropdown>
      ),
    },
    actionsRender: () => [
      <Space key="user">
        <span style={{ color: '#fff' }}>{user?.name}</span>
        <Avatar src={user?.avatar_url} size="small">
          {user?.name?.[0]}
        </Avatar>
      </Space>,
    ],
    token: {
      header: {
        colorBgHeader: '#001529',
        colorHeaderTitle: '#fff',
        colorTextMenu: 'rgba(255,255,255,0.75)',
        colorTextMenuSecondary: 'rgba(255,255,255,0.65)',
        colorTextMenuSelected: '#fff',
        colorBgMenuItemSelected: '#1890ff',
        colorTextMenuActive: '#fff',
        colorTextRightActionsItem: 'rgba(255,255,255,0.85)',
      },
      sider: {
        colorMenuBackground: '#fff',
        colorTextMenu: 'rgba(0,0,0,0.85)',
        colorTextMenuSelected: '#1890ff',
        colorBgMenuItemSelected: '#e6f7ff',
      },
    },
  };

  return (
    <ProLayout {...layoutProps}>
      <Outlet />
    </ProLayout>
  );
};

export default MainLayout;
