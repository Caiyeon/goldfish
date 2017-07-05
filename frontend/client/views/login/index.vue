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
                    <tr v-for="key in sessionKeys" v-if="key != 'token'">
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
import moment from 'moment'

export default {
  data () {
    return {
      type: 'Token',
      ID: '',
      Password: '',
      healthData: {},
      healthLoading: false
    }
  },

  mounted: function () {
    // demo
    this.$message({
      message: 'Try logging in with "goldfish" as token',
      type: 'warning',
      duration: 0,
      showCloseButton: true
    })
    // fetch vault cluster details
    this.getHealth()
  },

  computed: {
    session: function () {
      return this.$store.getters.session
    },
    healthKeys: function () {
      return Object.keys(this.healthData)
    },
    renewable: function () {
      return (this.session && this.session['renewable'])
    },
    sessionKeys: function () {
      return (this.session === null) || Object.keys(this.session)
    }
  },

  methods: {
    getHealth: function () {
      this.healthLoading = true
      this.$http.get('/api/health')
      .then((response) => {
        this.healthData = JSON.parse(response.data.result)
        this.healthData['server_time_utc'] = moment.utc(moment.unix(this.healthData['server_time_utc'])).format('ddd, h:mm:ss A MMMM Do YYYY') + ' GMT'
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
        headers: {'X-Vault-Token': this.session ? this.session.token : ''}
      })
      .then((response) => {
        // notify user, and clear inputs
        this.$notify({
          title: 'Login success!',
          message: '',
          type: 'success'
        })
        this.clearFormData()

        var newSession = {
          'token': response.data.result['cipher'],
          'type': this.type,
          'display_name': response.data.result['display_name'],
          'meta': response.data.result['meta'],
          'policies': response.data.result['policies'],
          'renewable': response.data.result['renewable'],
          'token_expiry': response.data.result['ttl'] === 0 ? 'never' : moment().add(response.data.result['ttl'], 'seconds').format('ddd, h:mm:ss A MMMM Do YYYY')
        }

        // store session data in localstorage and mutate vuex state
        window.localStorage.setItem('session', JSON.stringify(newSession))
        this.$store.commit('setSession', newSession)

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
        headers: {'X-Vault-Token': this.session ? this.session.token : ''}
      })
      .then((response) => {
        // deep copy session, update fields, and mutate state
        let newSession = JSON.parse(JSON.stringify(this.session))

        newSession['meta'] = response.data.result['meta']
        newSession['policies'] = response.data.result['policies']
        newSession['token_expiry'] = response.data.result['ttl'] === 0 ? 'never' : moment().add(response.data.result['ttl'], 'seconds').format('ddd, h:mm:ss A MMMM Do YYYY')

        window.localStorage.setItem('session', JSON.stringify(newSession))
        this.$store.commit('setSession', newSession)
        this.$notify({
          title: 'Renew success!',
          message: '',
          type: 'success'
        })
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
