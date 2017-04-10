import Vue from 'vue'
import axios from 'axios'
import NProgress from 'vue-nprogress'
import { sync } from 'vuex-router-sync'
import App from './App.vue'
import router from './router'
import store from './store'
import * as filters from './filters'
import { TOGGLE_SIDEBAR } from 'vuex-store/mutation-types'
import Notification from 'vue-bulma-notification'

Vue.prototype.$http = axios
Vue.axios = axios
Vue.use(NProgress)

// Enable devtools
Vue.config.devtools = true

sync(store, router)

const nprogress = new NProgress({ parent: '.nprogress-container' })

const { state } = store

router.beforeEach((route, redirect, next) => {
  if (state.app.device.isMobile && state.app.sidebar.opened) {
    store.commit(TOGGLE_SIDEBAR, false)
  }
  next()
})

Object.keys(filters).forEach(key => {
  Vue.filter(key, filters[key])
})

const app = new Vue({
  router,
  store,
  nprogress,
  ...App
})

const NotificationComponent = Vue.extend(Notification)

const openNotification = (propsData = {
  title: '',
  message: '',
  type: '',
  direction: '',
  duration: 4500,
  container: '.notifications'
}) => {
  return new NotificationComponent({
    el: document.createElement('div'),
    propsData
  })
}

Vue.prototype.$notify = openNotification

function handleError (error) {
  if (error.response.data.error) {
    openNotification({
      title: 'Error: ' + error.response.status,
      message: error.response.data.error,
      type: 'danger'
    })
    console.log(error.response.data.error)
  } else {
    openNotification({
      title: 'Error',
      message: 'Please login first',
      type: 'danger'
    })
    console.log(error.message)
  }
}

Vue.prototype.$onError = handleError

export { app, router, store }
