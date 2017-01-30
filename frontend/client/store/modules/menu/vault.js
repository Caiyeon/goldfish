import lazyLoading from './lazyLoading'

export default {
  meta: {
    label: 'Vault Admin',
    icon: 'fa-lock',
    expanded: true
  },
  children: [
    {
      name: 'Login',
      path: '/login',
      component: lazyLoading('vault/Login')
    },
    {
      name: 'Users',
      path: '/users',
      component: lazyLoading('vault/Users')
    }
  ]
}
