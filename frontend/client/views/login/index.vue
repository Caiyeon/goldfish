<template>
  <div>
    <div class="tile is-ancestor box is-vertical">
      <div class="tile">

        <!-- Left side -->
        <article class="tile is-parent is-5 is-vertical">

          <!-- Login tile -->
          <article class="tile is-child is-marginless is-paddingless">
            <h1 class="title">Vault Login</h1>
            <div class="box is-parent is-6">
              <form id="form" v-on:submit.prevent="login">

                <div class="control">
                  <label class="label">Authentication Type</label>
                  <div class="select is-fullwidth">
                    <select v-model="type" @change="clearFormData">
                      <option>Token</option>
                      <option>Userpass</option>
                    </select>
                  </div>
                </div>

                <!-- Token login form -->
                <p v-if="type === 'Token'" class="control has-icon">
                  <input class="input" type="password" placeholder="Vault Token" v-model="ID">
                  <span class="icon is-small">
                    <i class="fa fa-lock"></i>
                  </span>
                </p>

                <!-- Userpass login form -->
                <p v-if="type === 'Userpass'" class="control has-icon">
                  <input class="input" type="text" placeholder="Vault Username" v-model="ID">
                  <span class="icon is-small">
                    <i class="fa fa-user-circle-o"></i>
                  </span>
                </p>
                <p v-if="type === 'Userpass'" class="control has-icon">
                  <input class="input" type="password" placeholder="Vault Password" v-model="Password">
                  <span class="icon is-small">
                    <i class="fa fa-lock"></i>
                  </span>
                </p>

                <p class="control">
                  <button type="submit" value="Login" class="button is-primary">
                    Login
                  </button>
                </p>

              </form>
            </div>
          </article>

          <!-- Current session tile -->
          <article class="tile is-child is-marginless is-paddingless">
            <h1 class="title">Current Session</h1>
            <div class="box is-parent is-6">
              <div class="table-responsive">
                <table class="table is-striped is-narrow">
                  <thead>
                    <tr>
                      <th>Key</th>
                      <th>Value</th>
                    </tr>
                  </thead>
                  <tbody>
                    <tr v-for="key in sessionKeys">
                      <td>
                        {{ key }}
                      </td>
                      <td>
                        {{ sessionData[key] }}
                      </td>
                    </tr>
                  </tbody>
                </table>
                <p class="control">
                  <button class="button is-warning" @click="logout()">
                    Logout
                  </button>
                </p>
              </div>
            </div>
          </article>

        <!-- Left side (end) -->
        </article>


        <!-- Right side -->
        <article class="tile is-parent is-7 is-vertical">

          <!-- Vault Health Tile -->
          <article class="tile is-child is-marginless is-paddingless">
            <h1 class="title">Vault Health</h1>
            <div class="box is-parent is-6">
              <div class="table-responsive">
                <table class="table is-striped is-narrow">
                  <thead>
                    <tr>
                      <th>Key</th>
                      <th>Value</th>
                    </tr>
                  </thead>
                  <tbody>
                    <tr v-for="key in healthKeys">
                      <td>
                        {{ key }}
                      </td>
                      <td>
                        {{ healthData[key] }}
                      </td>
                    </tr>
                  </tbody>
                </table>
                <p class="control">
                  <button class="button is-primary"
                    v-bind:class="{
                      'is-loading': healthLoading,
                      'is-disabled': healthLoading
                    }"
                    @click="getHealth()">
                  Refresh
                </button>
              </p>
              </div>
            </div>
          </article>

        <!-- Right side (end) -->
        </article>

      </div>

    </div>
  </div>
</template>

<script>
  import Vue from 'vue'
  import Notification from 'vue-bulma-notification'

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

  export default {
    data () {
      return {
        csrf: '',
        type: 'Token',
        ID: '',
        Password: '',
        healthData: {},
        healthLoading: false,
        sessionData: {}
      }
    },

    mounted: function () {
      // fetch csrf for login post request later
      this.fetchCSRF()
      // fetch vault cluster details
      this.getHealth()
      var currentSession = window.localStorage.getItem('session')
      if (currentSession) {
        if (currentSession['cookie_expires_at'] > Date.now()) {
          window.localStorage.removeItem('session')
          openNotification({
            title: 'Session expired',
            message: 'Please login again',
            type: 'warning'
          })
        } else {
          this.sessionData = JSON.parse(currentSession)
        }
      }
    },

    computed: {
      healthKeys: function () {
        return Object.keys(this.healthData)
      },
      sessionKeys: function () {
        return Object.keys(this.sessionData)
      }
    },

    methods: {
      fetchCSRF: function () {
        this.$http.get('/api/login/csrf')
          .then((response) => {
            this.csrf = response.headers['x-csrf-token']
          })
          .catch((error) => {
            handleError(error)
          })
      },

      getHealth: function () {
        this.healthLoading = true
        this.$http.get('/api/health')
          .then((response) => {
            this.healthData = JSON.parse(response.data.result)
            this.healthData['server_time_utc'] = new Date(this.healthData['server_time_utc'] * 1000).toUTCString()
            this.healthLoading = false
          })
          .catch((error) => {
            handleError(error)
            this.healthLoading = false
          })
      },

      login: function () {
        this.$http
          .post('/api/login', {
            Type: this.type.toLowerCase(),
            ID: this.ID,
            Password: this.Password
          }, {
            headers: {'X-CSRF-Token': this.csrf}
          })

          .then((response) => {
            // notify user, and clear inputs
            openNotification({
              title: 'Login success!',
              message: '',
              type: 'success'
            })
            this.clearFormData()

            // set user's session reactively, and store it browser's localStorage
            this.sessionData = {
              'type': this.type,
              'display_name': response.data.data['display_name'],
              'meta': response.data.data['meta'],
              'policies': response.data.data['policies'],
              'renewable': response.data.data['renewable'],
              'token_expiry': response.data.data['ttl'] === 0 ? 'never' : new Date(Date.now() + response.data.data['ttl'] * 1000).toLocaleString(),
              'cookie_expiry': new Date(Date.now() + 28800000).toLocaleString() // 8 hours from now
            }
            window.localStorage.setItem('session', JSON.stringify(this.sessionData))
          })

          .catch((error) => {
            handleError(error)
          })
      },

      logout: function () {
        document.cookie = 'auth=; Path=/; Expires=Thu, 01 Jan 1970 00:00:01 GMT;'
        this.sessionData = {}
        window.localStorage.removeItem('session')
      },

      clearFormData: function () {
        this.ID = ''
        this.Password = ''
      }
    }

  }
</script>

<style scoped>
  .button {
    margin: 5px 0 0;
  }

  .control .button {
    margin: inherit;
  }
</style>
