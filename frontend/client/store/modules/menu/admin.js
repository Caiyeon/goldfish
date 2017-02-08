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
      name: 'Health',
      path: '/health',
      component: lazyLoading('admin/Health')
    }
  ]
}
