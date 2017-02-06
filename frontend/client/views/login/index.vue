<template>
  <div>
    <div class="tile is-ancestor">

      <div id="loginapp" class="tile is-parent is-6">
        <article class="tile is-child box">
          <h1 class="title">Vault Login</h1>
          <div class="block">
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
                <i class="fa fa-lock"></i>
              </p>
              <p class="control">
                <button type="submit" value="Login" class="button is-success">
                  Login
                </button>
              </p>

            </form>
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
        type: 'Token',
        vaultToken: '',
        csrf: ''
      }
    },
    mounted: function () {
      this.$http.get('/api/login/csrf').then(function (response) {
        this.csrf = response.headers.get('x-csrf-token')
      }, function (err) {
        console.log(err.body.error)
      })
    },
    methods: {
      login: function () {
        var payload = {
          Type: this.type.toLowerCase(),
          ID: this.vaultToken
        }
        var headers = {
          headers: {
            'X-CSRF-Token': this.csrf
          }
        }
        this.$http.post('/api/login', payload, headers).then(function (response) {
          console.log(response.data.status)
          openNotification({
            title: 'Login success!',
            message: '',
            type: 'success'
          })
          this.vaultToken = ''
        }, function (err) {
          console.log(err.body.error)
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
