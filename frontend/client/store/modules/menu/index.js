import * as types from '../../mutation-types'
import lazyLoading from './lazyLoading'
import admin from './admin'
import tools from './tools'

// defaults
// import charts from './charts'
// import uifeatures from './uifeatures'
// import components from './components'
// import tables from './tables'

// show: meta.label -> name
// name: component name
// meta.label: display label

const state = {
  items: [
    {
      name: 'Login',
      path: '/login',
      meta: {
        icon: 'fa-lock'
      },
      component: lazyLoading('login', true)
    },
    {
      name: 'Secrets',
      path: '/secrets',
      meta: {
        icon: 'fa-list'
      },
      component: lazyLoading('secrets', true)
    },
    {
      name: 'Bulletin',
      path: '/bulletinboard',
      meta: {
        icon: 'fa-thumb-tack'
      },
      component: lazyLoading('bulletinboard', true)
    },
    admin,
    tools
  ]
}

const mutations = {
  [types.EXPAND_MENU] (state, menuItem) {
    if (menuItem.index > -1) {
      if (state.items[menuItem.index] && state.items[menuItem.index].meta) {
        state.items[menuItem.index].meta.expanded = menuItem.expanded
      }
    } else if (menuItem.item && 'expanded' in menuItem.item.meta) {
      menuItem.item.meta.expanded = menuItem.expanded
    }
  }
}

export default {
  state,
  mutations
}
