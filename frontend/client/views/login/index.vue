<template>
  <div>
    <div class="tile is-ancestor box is-vertical">
      <div class="tile">

        <!-- Left side -->
        <article class="tile is-parent is-5 is-vertical">

          <!-- Login tile -->
          <article class="tile is-child is-marginless is-paddingless">
            <h1 class="title">Vault Login</h1>
            <div class="box is-parent is-6" @keyup.enter="login">

              <div class="field">
                <div class="control">
                  <label class="label">Authentication Type</label>
                  <div class="select is-fullwidth">
                    <select v-model="type" @change="clearFormData">
                      <option v-bind:value="'Token'">Token</option>
                      <option v-bind:value="'Userpass'">Userpass</option>
                      <option v-bind:value="'Github'">Github</option>
                      <option v-bind:value="'LDAP'">LDAP</option>
                    </select>
                  </div>
                </div>
              </div>

              <!-- Token login form -->
              <div v-if="type === 'Token'" class="field">
                <p class="control has-icons-left">
                  <input class="input" type="password" placeholder="Vault Token" v-model="ID">
                  <span class="icon is-small">
                    <i class="fa fa-lock"></i>
                  </span>
                </p>
              </div>

              <!-- Userpass login form -->
              <div v-if="type === 'Userpass'" class="field">
                <div class="field">
                  <p class="control has-icons-left">
                    <input class="input" type="text" placeholder="Vault Username" v-model="ID">
                    <span class="icon is-small">
                      <i class="fa fa-user-circle-o"></i>
                    </span>
                  </p>
                </div>
                <div class="field">
                  <p class="control has-icons-left">
                    <input class="input" type="password" placeholder="Vault Password" v-model="Password">
                    <span class="icon is-small">
                      <i class="fa fa-lock"></i>
                    </span>
                  </p>
                </div>
              </div>

              <!-- Github login form -->
              <div v-if="type === 'Github'" class="field">
                <p class="control has-icons-left">
                  <input class="input" type="password" placeholder="Github Access Token" v-model="ID">
                  <span class="icon is-small">
                    <i class="fa fa-lock"></i>
                  </span>
                </p>
              </div>

              <!-- LDAP login form -->
              <div v-if="type === 'LDAP'" class="field">
                <div class="field">
                  <p class="control has-icons-left">
                    <input class="input" type="text" placeholder="Username" v-model="ID">
                    <span class="icon is-small">
                      <i class="fa fa-user-circle-o"></i>
                    </span>
                  </p>
                </div>
                <div class="field">
                  <p class="control has-icons-left">
                    <input class="input" type="password" placeholder="Password" v-model="Password">
                    <span class="icon is-small">
                      <i class="fa fa-lock"></i>
                    </span>
                  </p>
                </div>
              </div>

              <div class="field">
                <p class="control">
                  <button @click="login" type="submit" value="Login" class="button is-primary">
                    Login
                  </button>
                </p>
              </div>

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
                        {{ session[key] }}
                      </td>
                    </tr>
                  </tbody>
                </table>
                <p v-if="session !== null" class="control">
                  <button class="button is-warning" @click="logout()">
                    Logout
                  </button>
                  <button v-if="renewable" class="button is-primary"
                  @click="renewLogin()">
                    Renew
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
export default {
  data () {
    return {
      csrf: '',
      type: 'Token',
      ID: '',
      Password: '',
      healthData: {},
      healthLoading: false
    }
  },

  mounted: function () {
    // fetch csrf for login post request later
    this.fetchCSRF()
    // fetch vault cluster details
    this.getHealth()
  },

  computed: {
    healthKeys: function () {
      return Object.keys(this.healthData)
    },
    renewable: function () {
      return (this.session && this.session['renewable'])
    },
    session: function () {
      return this.$store.getters.session
    },
    sessionKeys: function () {
      return (this.session === null) || Object.keys(this.session)
    }
  },

  methods: {
    fetchCSRF: function () {
      this.$http.get('/api/login/csrf')
      .then((response) => {
        this.csrf = response.headers['x-csrf-token']
      })
      .catch((error) => {
        this.$onError(error)
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
        this.$onError(error)
        this.healthLoading = false
      })
    },

    login: function () {
      this.$http.post('/api/login', {
        Type: this.type.toLowerCase(),
        ID: this.ID,
        Password: this.Password
      }, {
        headers: {'X-CSRF-Token': this.csrf}
      })
      .then((response) => {
        // notify user, and clear inputs
        this.$notify({
          title: 'Login success!',
          message: '',
          type: 'success'
        })
        this.clearFormData()

        // construct session data
        this.session = {
          'type': this.type,
          'display_name': response.data.data['display_name'],
          'meta': response.data.data['meta'],
          'policies': response.data.data['policies'],
          'renewable': response.data.data['renewable'],
          'token_expiry': response.data.data['ttl'] === 0 ? 'never' : new Date(Date.now() + response.data.data['ttl'] * 1000).toString(),
          'cookie_expiry': new Date(Date.now() + 28800000).toString() // 8 hours from now
        }

        // store session data in localstorage
        window.localStorage.setItem('session', JSON.stringify(this.session))

        // mutate state of vuex
        this.$store.commit('setSession', this.session)

        // notify user of generated client-token
        if (this.type === 'Userpass' || this.type === 'LDAP') {
          this.$message({
            message: 'Your access token is: ' + response.data.data['id'] + ' and this is the only time you will see it. If you wish, you may login with this to avoid creating unnecessary access tokens in the future.',
            type: 'warning',
            duration: 0,
            showCloseButton: true
          })
        }
      })
      .catch((error) => {
        this.$onError(error)
      })
    },

    logout: function () {
      // force cookie timeout
      document.cookie = 'auth=; Path=/; Expires=Thu, 01 Jan 1970 00:00:01 GMT;'
      // purge session from localstorage
      window.localStorage.removeItem('session')
      // mutate vuex state
      this.$store.commit('clearSession')
    },

    clearFormData: function () {
      this.ID = ''
      this.Password = ''
    },

    renewLogin: function () {
      this.$http.post('/api/login/renew-self', {}, {
        headers: {'X-CSRF-Token': this.csrf}
      })
      .then((response) => {
        this.$notify({
          title: 'Renew success!',
          message: '',
          type: 'success'
        })
        this.session['meta'] = response.data.data['meta']
        this.session['policies'] = response.data.data['policies']
        this.session['token_expiry'] = response.data.data['ttl'] === 0 ? 'never' : new Date(Date.now() + response.data.data['ttl'] * 1000).toString()

        // store session data in localstorage
        window.localStorage.setItem('session', JSON.stringify(this.session))

        // mutate state of vuex
        this.$store.commit('setSession', this.session)
      })
      .catch((error) => {
        this.$onError(error)
      })
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
