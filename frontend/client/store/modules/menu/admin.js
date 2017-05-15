import lazyLoading from './lazyLoading'

export default {
  meta: {
    label: 'Administration',
    icon: 'fa-cogs',
    expanded: true
  },
  children: [
    {
      name: 'Users',
      path: '/users',
      component: lazyLoading('admin/Users')
    },
    {
      name: 'Policies',
      path: '/policies',
      component: lazyLoading('admin/Policies')
    },
    {
      name: 'Mounts',
      path: '/mounts',
      component: lazyLoading('admin/Mounts')
    },
    {
      name: 'Requests',
      path: '/requests',
      component: lazyLoading('admin/Requests')
    }
  ]
}
