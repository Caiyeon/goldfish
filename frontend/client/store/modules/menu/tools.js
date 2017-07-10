import lazyLoading from './lazyLoading'

export default {
  meta: {
    label: 'Tools',
    icon: 'fa-wrench',
    expanded: true
  },
  children: [
    {
      name: 'Transit',
      path: '/transit',
      component: lazyLoading('tools/Transit')
    },
    {
      name: 'Token Creator',
      path: '/create-token',
      component: lazyLoading('tools/CreateToken')
    },
    {
      name: 'Wrapper',
      path: '/wrapper',
      component: lazyLoading('tools/Wrapper')
    },
    {
      name: 'Dependencies',
      path: '/dependencies',
      component: lazyLoading('tools/Dependencies')
    }
  ]
}
