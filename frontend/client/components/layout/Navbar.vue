<template>
  <section class="hero is-bold app-navbar animated" :class="{ slideInDown: show, slideOutDown: !show }">
    <div class="hero-head">

      <nav class="nav">

        <div class="nav-left">
          <a class="nav-item is-hidden-tablet" @click="toggleSidebar(!sidebar.opened)">
            <i class="fa fa-bars" aria-hidden="true"></i>
          </a>
        </div>

        <div class="nav-center">
          <a class="nav-item hero-brand" href="/">
            <img src="~assets/logo.svg" :alt="pkginfo.description">
            <tooltip :label="'v' + pkginfo.version" placement="right" type="success" size="small" :no-animate="true" :always="true" :rounded="true">
              <div class="is-hidden-mobile">
                <span class="vue">Gold</span><strong class="admin">fish</strong>
              </div>
            </tooltip>
          </a>
        </div>

        <div class="nav-right is-flex">

          <!-- session dropdown -->
          <div v-if="session" class="nav-item">
            <dropdown animation="ani-slide-y"
            :position="position"
            :visible="profileDropdown">

              <!-- profile button -->
              <a class="button is-primary is-outlined"
              @click="profileDropdown = !profileDropdown">
                <span class="icon">
                  <i class="fa fa-user"></i>
                </span>
                <span class="is-hidden-mobile">Session</span>
              </a>

              <!-- dropdown menu -->
              <div slot="dropdown" class="dialog">
                <div class="box">
                <aside class="menu">
                  <p class="menu-label">
                    token-public-demo-token
                  </p>
                  <ul class="menu-list">
                    <li v-if="tokenExpiresIn === ''">Token will never expire</li>
                    <li v-else>Token expires {{tokenExpiresIn}}</li>
                    <li>Cookie expires {{cookieExpiresIn}}</li>
                  </ul>
                </aside>
                </div>
              </div>

            </dropdown>
          </div>

          <!-- github button -->
          <div class="nav-item">
            <div class="field is-grouped">
              <p class="control">
                <a class="button is-info is-outlined"
                href="https://github.com/Caiyeon/goldfish">
                  <span class="icon">
                    <i class="fa fa-github"></i>
                  </span>
                  <span class="is-hidden-mobile">Source Code</span>
                </a>
              </p>
            </div>
          </div>

        </div>

      </nav>

    </div>
  </section>
</template>

<script>
import Tooltip from 'vue-bulma-tooltip'
import { mapGetters, mapActions } from 'vuex'
import dropdown from 'vue-my-dropdown'
import moment from 'moment'

export default {

  components: {
    Tooltip,
    dropdown
  },

  props: {
    show: Boolean
  },

  data () {
    return {
      profileDropdown: false,
      position: ['center', 'bottom', 'center', 'top'],
      now: moment()
    }
  },

  mounted: function () {
    setInterval(() => {
      this.now = moment()
    }, 1000)

    // if session cookie is still valid, load session data
    let raw = window.localStorage.getItem('session')
    if (raw) {
      var session = JSON.parse(raw)
      if (Date.now() > Date.parse(session['cookie_expiry'])) {
        window.localStorage.removeItem('session')
        this.$notify({
          title: 'Session expired',
          message: 'Please login again',
          type: 'warning'
        })
        this.$store.commit('clearSession')
      } else {
        this.$store.commit('setSession', session)
      }
    } else {
      this.$store.commit('clearSession')
    }
    // uncomment this to see the details of the session everytime you refresh the page
    // console.log(JSON.stringify(this.session))
  },

  computed: {
    ...mapGetters({
      session: 'session',
      pkginfo: 'pkg',
      sidebar: 'sidebar'
    }),

    tokenExpiresIn: function () {
      if (this.session === null) {
        return ''
      }
      if (this.session['token_expiry'] === 'never') {
        return ''
      }
      return this.now.to(moment(this.session['token_expiry'], 'ddd, h:mm:ss A'))
    },

    cookieExpiresIn: function () {
      if (this.session === null) {
        return ''
      }
      return this.now.to(moment(this.session['cookie_expiry'], 'ddd, h:mm:ss A'))
    }
  },

  methods: {
    ...mapActions([
      'toggleSidebar'
    ])
  }
}
</script>

<style lang="scss">
@import '~bulma/sass/utilities/variables';

.app-navbar {
  position: fixed;
  min-width: 100%;
  z-index: 1024;
  box-shadow: 0 2px 3px rgba(17, 17, 17, 0.1), 0 0 0 1px rgba(17, 17, 17, 0.1);

  .container {
    margin: auto 10px;
  }

  .nav-right {
    align-items: stretch;
    align-items: stretch;
    flex: 1;
    justify-content: flex-end;
    overflow: hidden;
    overflow-x: auto;
    white-space: nowrap;
  }
}

.hero-brand {
  .vue {
    margin-left: 10px;
    color: #36AC70;
  }
  .admin {
    color: #28374B;
  }
}

/* If not anchored, dropdown will move down as page scrolls */
.my-dropdown-dd {
  top: 47px !important;
}
</style>
