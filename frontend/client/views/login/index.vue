<template>
  <div>
    <div class="tile is-ancestor">

      <div class="tile is-parent box">
        <article class="tile is-parent is-child is-5">
          <h1 class="title">Vault Login</h1>
          <div class="box is-parent is-6">
            <form id="form" v-on:submit.prevent="login">

              <div class="control">
                <label class="label">Authentication Type</label>
                <div class="select is-fullwidth">
                  <select v-model="type">
                    <option>Token</option>
                  </select>
                </div>
              </div>

              <p class="control has-icon">
                <input class="input" type="password" placeholder="Vault Token" v-model="vaultToken">
                <span class="icon is-small">
                  <i class="fa fa-lock"></i>
                </span>
              </p>
              <p class="control">
                <button type="submit" value="Login" class="button is-success">
                  Login
                </button>
              </p>

            </form>
          </div>
        </article>

        <article class="tile is-parent is-child is-7">
          <h1 class="title">Vault Health</h1>
          <div class="box">
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
            </div>
          </div>
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

  export default {
    data () {
      return {
        csrf: '',
        type: 'Token',
        vaultToken: '',
        healthData: {}
      }
    },
    mounted: function () {
      // fetch csrf for login post request later
      this.$http.get('/api/login/csrf')
        .then((response) => {
          this.csrf = response.headers['x-csrf-token']
        })
        .catch((error) => {
          openNotification({
            title: 'Error',
            message: error.response.data,
            type: 'danger'
          })
          console.log(error.response.data)
        })

      // fetch vault cluster details
      this.$http.get('/api/health')
        .then((response) => {
          this.healthData = JSON.parse(response.data.result)
        })
        .catch((error) => {
          openNotification({
            title: 'Error',
            message: error.response.data,
            type: 'danger'
          })
          console.log(error.response.data)
        })
    },

    computed: {
      healthKeys: function () {
        return Object.keys(this.healthData)
      }
    },

    methods: {
      login: function () {
        this.$http
          .post('/api/login', {
            Type: this.type.toLowerCase(),
            ID: this.vaultToken
          }, {
            headers: {'X-CSRF-Token': this.csrf}
          })
          .then((response) => {
            openNotification({
              title: 'Login success!',
              message: '',
              type: 'success'
            })
            this.vaultToken = ''
          })
          .catch((error) => {
            openNotification({
              title: 'Error',
              message: error.response.data,
              type: 'danger'
            })
            console.log(error.response.data)
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
