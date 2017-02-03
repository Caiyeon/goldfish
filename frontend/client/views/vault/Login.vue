<template>
  <div>
    <div class="tile is-ancestor">

      <div id="loginapp" class="tile is-parent is-6">
        <article class="tile is-child box">
          <h1 class="title">Vault Login</h1>
          <div class="block">
            <form id="form" v-on:submit.prevent="login">
              <p class="control has-icon">
                <input class="input" type="url" placeholder="http://127.0.0.1:8200" v-model="vaultAddress">
                <i class="fa fa-location-arrow"></i>
              </p>
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

  export default {
    data () {
      return {
        vaultAddress: '',
        vaultToken: '',
        statusText: ''
      }
    },
    created: function () {
      // fetch CSRF token and push it as an interceptor for POST request upon login
      this.$http.get('/api/login/csrf').then(function (response) {
        Vue.http.interceptors.push((request, next) => {
          request.headers.set('X-CSRF-Token', response.headers.get('x-csrf-token'))
          next()
        })
      }, function (err) {
        this.statusText = err.statusText
        console.log(err.statusText)
      })
    },
    methods: {
      login: function () {
        this.statusText = ''
        var payload = {
          addr: this.vaultAddress,
          token: this.vaultToken
        }
        this.$http.post('/api/login', payload).then(function (response) {
          console.log(response.data.status)
          this.$router.push({
            name: 'Users'
          })
          this.$router.go(0)
        }, function (err) {
          this.statusText = err.statusText
          console.log(err.statusText)
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
