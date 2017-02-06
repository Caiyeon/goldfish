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
    }
  ]
}
