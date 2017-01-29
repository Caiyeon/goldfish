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
      icon: 'fa-lock',
      component: lazyLoading('vault/Buttons')
    }
  ]
}
