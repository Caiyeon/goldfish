import lazyLoading from './lazyLoading'

export default {
  meta: {
    label: 'Vault Admin',
    icon: 'fa-lock',
    expanded: true
  },

  children: [
    {
      name: 'Buttons',
      path: '/buttons2',
      component: lazyLoading('vault/Buttons')
    },
    {
      name: 'Login',
      path: '/login',
      component: lazyLoading('vault/Login')
    }
  ]
}
