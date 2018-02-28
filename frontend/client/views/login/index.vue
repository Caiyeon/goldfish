<template>
  <div>
    <div class="tile is-ancestor box is-vertical">
      <div class="tile">

        <!-- Left side -->
        <article class="tile is-parent is-5 is-vertical">

          <!-- Bootstrap tile -->
          <article v-if="goldfishHealthData && goldfishHealthData['bootstrapped'] === false"
            class="tile is-child is-marginless is-paddingless">
            <h2 class="title is-3">Welcome!</h2>

            <div class="box is-parent is-6">
              <label class="label">Setting up Goldfish</label>
              <div class="field has-addons">
                <div class="control">
                  <input class="input" type="text" v-model="secretID"
                  placeholder="Insert wrapping token" @keyup.enter="bootstrapGoldfish()">
                  <p class="help is-info">
                    vault write -f -wrap-ttl=5m auth/approle/role/goldfish/secret-id
                  </p>
                </div>
                <div class="control">
                  <button class="button is-info"
                  v-bind:class="{ 'is-loading': bootstrapLoading }"
                  :disabled="secretID === ''"
                  @click="bootstrapGoldfish()">
                    Swim!
                  </button>
                </div>
              </div>
            </div>
          </article>

          <!-- Login tile -->
          <article class="tile is-child is-marginless is-paddingless">
            <h2 class="subtitle is-4">Vault Login</h2>
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
                      <option v-bind:value="'Okta'">Okta</option>
                    </select>
                  </div>
                </div>
              </div>

              <!-- Custom login path -->
              <div v-if="bCustomPath && type !== 'Token'" class="field">
                <p class="control has-icons-left">
                  <input class="input" type="text" placeholder="Mount name e.g. 'ldap2'" v-model="customPath">
                  <span class="icon is-small is-left">
                    <i class="fa fa-tasks"></i>
                  </span>
                </p>
              </div>

              <!-- Token login form -->
              <div v-if="type === 'Token'" class="field">
                <p class="control has-icons-left">
                  <input class="input" type="password" placeholder="Vault Token" v-model="ID">
                  <span class="icon is-small is-left">
                    <i class="fa fa-lock"></i>
                  </span>
                </p>
              </div>

              <!-- Userpass login form -->
              <div v-if="type === 'Userpass'" class="field">
                <div class="field">
                  <p class="control has-icons-left">
                    <input class="input" type="text" placeholder="Vault Username" v-model="ID">
                    <span class="icon is-small is-left">
                      <i class="fa fa-user-circle-o"></i>
                    </span>
                  </p>
                </div>
                <div class="field">
                  <p class="control has-icons-left">
                    <input class="input" type="password" placeholder="Vault Password" v-model="password">
                    <span class="icon is-small is-left">
                      <i class="fa fa-lock"></i>
                    </span>
                  </p>
                </div>
              </div>

              <!-- Github login form -->
              <div v-if="type === 'Github'" class="field">
                <p class="control has-icons-left">
                  <input class="input" type="password" placeholder="Github Access Token" v-model="ID">
                  <span class="icon is-small is-left">
                    <i class="fa fa-lock"></i>
                  </span>
                </p>
              </div>

              <!-- LDAP login form -->
              <div v-if="type === 'LDAP'" class="field">
                <div class="field">
                  <p class="control has-icons-left">
                    <input class="input" type="text" placeholder="Username" v-model="ID">
                    <span class="icon is-small is-left">
                      <i class="fa fa-user-circle-o"></i>
                    </span>
                  </p>
                </div>
                <div class="field">
                  <p class="control has-icons-left">
                    <input class="input" type="password" placeholder="Password" v-model="password">
                    <span class="icon is-small is-left">
                      <i class="fa fa-lock"></i>
                    </span>
                  </p>
                </div>
              </div>

              <!-- Okta login form -->
              <div v-if="type === 'Okta'" class="field">
                <div class="field">
                  <p class="control has-icons-left">
                    <input class="input" type="text" placeholder="Username" v-model="ID">
                    <span class="icon is-small is-left">
                      <i class="fa fa-user-circle-o"></i>
                    </span>
                  </p>
                </div>
                <div class="field">
                  <p class="control has-icons-left">
                    <input class="input" type="password" placeholder="Password" v-model="password">
                    <span class="icon is-small is-left">
                      <i class="fa fa-lock"></i>
                    </span>
                  </p>
                </div>
              </div>

              <div v-if="type !== 'Token'" class="field">
                <div class="control">
                  <label class="checkbox">
                    <input type="checkbox" v-model="bCustomPath">
                    Custom path
                  </label>
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
            <h2 class="subtitle is-4">Current Session</h2>
            <div class="box is-parent is-6">
              <table class="table is-fullwidth is-striped is-narrow">
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
                    <td v-if="key === 'policies'">
                      <div class="tags">
                        <span v-for="policy in session['policies']" class="tag is-rounded is-info">
                          {{policy}}
                        </span>
                      </div>
                    </td>
                    <td v-else>
                      {{ session[key] }}
                    </td>
                  </tr>
                </tbody>
              </table>
              <p v-if="session !== null" class="control">
                <button v-if="renewable" class="button is-primary"
                @click="renewLogin()">
                  Renew
                </button>
                <button class="button is-warning" @click="logout(false)">
                  Logout
                </button>
                <button class="button is-warning" @click="logout(true)">
                  Revoke Token
                </button>
              </p>
            </div>
          </article>

        <!-- Left side (end) -->
        </article>


        <!-- Right side -->
        <article class="tile is-parent is-7 is-vertical">

          <!-- Vault Health Tile -->
          <article class="tile is-child is-marginless is-paddingless">
            <h2 class="subtitle is-4">Vault Health</h2>
            <div class="box is-parent is-6">
              <table class="table is-fullwidth is-striped is-narrow">
                <thead>
                  <tr>
                    <th>Key</th>
                    <th>Value</th>
                  </tr>
                </thead>
                <tbody>
                  <tr v-for="key in Object.keys(vaultHealthData)">
                    <td>
                      {{ key }}
                    </td>
                    <td>
                      {{ vaultHealthData[key] }}
                    </td>
                  </tr>
                </tbody>
              </table>
              <p class="control">
                <button class="button is-primary"
                  v-bind:class="{ 'is-loading': vaultHealthLoading }"
                  :disabled="vaultHealthLoading"
                  @click="getVaultHealth()">
                Refresh
                </button>
              </p>
            </div>
          </article>

          <!-- Goldfish Health tile -->
          <article class="tile is-child is-marginless is-paddingless">
            <h2 class="subtitle is-4">Goldfish Health</h2>
            <div class="box is-parent is-6">
              <table class="table is-fullwidth is-striped is-narrow">
                <thead>
                  <tr>
                    <th>Key</th>
                    <th>Value</th>
                  </tr>
                </thead>
                <tbody>
                  <tr v-for="key in Object.keys(goldfishHealthData)">
                    <td>
                      {{ key }}
                    </td>
                    <td>
                      {{ goldfishHealthData[key] }}
                    </td>
                  </tr>
                </tbody>
              </table>
              <p class="control">
                <button class="button is-primary"
                  v-bind:class="{ 'is-loading': goldfishHealthLoading }"
                  :disabled="goldfishHealthLoading"
                  @click="getGoldfishHealth()">
                Refresh
                </button>
              </p>
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
      password: '',
      vaultHealthData: {},
      vaultHealthLoading: false,
      goldfishHealthData: {},
      goldfishHealthLoading: false,
      secretID: '',
      bootstrapLoading: false,
      bCustomPath: false,
      customPath: ''
    }
  },

  mounted: function () {
    this.getVaultHealth()
    this.getGoldfishHealth()
  },

  computed: {
    session: function () {
      return this.$store.getters.session
    },
    renewable: function () {
      return (this.session && this.session['renewable'])
    },
    sessionKeys: function () {
      return (this.session === null) || Object.keys(this.session)
    }
  },

  methods: {
    bootstrapGoldfish: function () {
      this.bootstrapLoading = true
      this.$http.post('/v1/bootstrap', {
        wrapping_token: this.secretID
      })
      .then((response) => {
        this.$notify({
          title: 'Success',
          message: 'Goldfish successfully bootstrapped!',
          type: 'success'
        })
        this.secretID = ''
        this.bootstrapLoading = false
        // reload health so that the bootstrap tile can be toggled off by vue
        this.getGoldfishHealth()
      })
      .catch((error) => {
        this.secretID = ''
        this.bootstrapLoading = false
        this.$onError(error)
      })
    },

    getVaultHealth: function () {
      this.vaultHealthLoading = true
      this.$http.get('/v1/vaulthealth')
      .then((response) => {
        this.vaultHealthData = response.data.result
        this.vaultHealthData['server_time_utc'] = moment.utc(
          moment.unix(this.vaultHealthData['server_time_utc']))
          .format('ddd, h:mm:ss A MMMM Do YYYY') + ' GMT'
        this.vaultHealthLoading = false
      })
      .catch((error) => {
        this.$onError(error)
        this.vaultHealthLoading = false
      })
    },

    getGoldfishHealth: function () {
      this.goldfishHealthLoading = true
      this.$http.get('/v1/health')
      .then((response) => {
        this.goldfishHealthData = response.data
        this.goldfishHealthData['deployment_time_utc'] = moment.utc(
          moment.unix(this.goldfishHealthData['deployment_time_utc']))
          .format('ddd, h:mm:ss A MMMM Do YYYY') + ' GMT'
        this.goldfishHealthLoading = false
      })
      .catch((error) => {
        if (error.response.data.error === 'Vault:  permission denied') {
          this.$notify({
            title: 'Error',
            message: 'Goldfish server could not authenticate against vault',
            type: 'danger'
          })
        } else {
          this.$onError(error)
        }
        this.goldfishHealthLoading = false
      })
    },

    login: function () {
      this.$http.post('/v1/login', {
        type: this.type.toLowerCase(),
        id: this.ID,
        password: this.password,
        path: this.bCustomPath ? this.customPath.trim('/') : ''
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
        if (this.type === 'Userpass' || this.type === 'LDAP' || this.type === 'Okta') {
          this.$message({
            message: 'Your access token is: ' + response.data.result['id'] + ' and this is the only time you will see it. If you wish, you may login with this to avoid creating unnecessary access tokens in the future.',
            type: 'warning',
            duration: 0,
            showCloseButton: true
          })
        }
      })
      .catch((error) => {
        // to avoid ambiguity, current session should be purged when new login fails
        this.logout(false)
        this.$onError(error)
        if (this.bCustomPath && error.response.status === 400 &&
          error.response.data.error === 'Vault:  missing client token') {
          this.$notify({
            title: 'Custom path?',
            message: 'If the custom path does not exist, vault will respond with error 400',
            type: 'warning',
            duration: 10000
          })
        }
      })
    },

    logout: function (revoke) {
      // if user wants to revoke token
      if (revoke) {
        this.$http.post('/v1/token/revoke-self', {}, {
          headers: {'X-Vault-Token': this.session ? this.session.token : ''}
        })
        .then((response) => {
          // notify user, and clear inputs
          this.$notify({
            title: 'Token revoked!',
            message: '',
            type: 'success'
          })
          // purge session from localstorage
          window.localStorage.removeItem('session')
          // mutate vuex state
          this.$store.commit('clearSession')
        })
        .catch((error) => {
          this.$onError(error)
        })
      }
    },

    clearFormData: function () {
      this.ID = ''
      this.password = ''
      this.bCustomPath = false
      this.customPath = ''
    },

    renewLogin: function () {
      this.$http.post('/v1/login/renew-self', {}, {
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
