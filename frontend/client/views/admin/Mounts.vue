<template>
  <div>
    <div class="tile is-ancestor">

      <div class="tile is-parent is-vertical is-6">
        <article class="tile is-child box">
          <div class="table-responsive">
            <table class="table is-striped is-narrow">
              <thead>
                <tr>
                  <th>Type</th>
                  <th>Path</th>
                  <th>Def_TTL</th>
                  <th>Max_TTL</th>
                </tr>
              </thead>
              <tbody>
                <tr v-for="(mount, index) in mounts">
                  <td class="is-icon">
                    <a class="tag is-danger is-disabled is-pulled-left">
                      {{ mount.type }}
                    </a>
                  </td>
                  <td>
                    <tooltip v-bind:label="mount.desc" placement="right" type="info" :rounded="true" >
                      <a @click="getMountConfig(index)">
                        {{ mount.path }}
                      </a>
                    </tooltip>
                  </td>
                  <td class="is-icon">
                    <a class="tag is-primary is-disabled is-pulled-left">
                      {{ mount.conf.default_lease_ttl === 0 ? 'Default' : mount.conf.default_lease_ttl }}
                    </a>
                  </td>
                  <td class="is-icon">
                    <a class="tag is-primary is-disabled is-pulled-left">
                      {{ mount.conf.max_lease_ttl === 0 ? 'Default' : mount.conf.max_lease_ttl }}
                    </a>
                  </td>
                </tr>
              </tbody>
            </table>
          </div>
        </article>
      </div>

      <div class="tile is-parent is-vertical is-6">
        <article class="tile is-child box">
          <h4 class="title is-4">Mount Config</h4>
          <p class="control">
            <textarea class="textarea" placeholder="Select a mount" v-model="mountConfigModified"></textarea>
          </p>
          <p class="control is-pulled-right">
            <a @click="postMountConfig"
            class="button is-primary is-outlined"
            v-bind:class="{ 'is-disabled': (mountConfig === mountConfigModified) }">
              Submit Changes
              <span class="icon is-small">
                <i class="fa fa-check"></i>
              </span>
            </a>
          </p>
        </article>
      </div>

    </div>

  </div>
</template>

<script>
  import Vue from 'vue'
  import Notification from 'vue-bulma-notification'
  import Tooltip from 'vue-bulma-tooltip'

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
    components: {
      Tooltip
    },

    data () {
      return {
        csrf: '',
        mounts: [],
        mountConfig: '',
        mountConfigModified: '',
        selectedIndex: -1
      }
    },

    computed: {
    },

    filters: {
    },

    mounted: function () {
      this.$http.get('/api/mounts').then(function (response) {
        this.mounts = []
        this.csrf = response.headers.get('x-csrf-token')
        var keys = Object.keys(response.data.result)
        for (var i = 0; i < keys.length; i++) {
          this.mounts.push({
            path: keys[i],
            type: response.data.result[keys[i]]['type'],
            desc: response.data.result[keys[i]]['description'],
            conf: response.data.result[keys[i]]['config']
          })
        }
      }, function (err) {
        openNotification({
          title: 'Error',
          message: err.body.error,
          type: 'danger'
        })
        console.log(err.body.error)
      })
    },

    methods: {
      getMountConfig: function (index) {
        this.selectedIndex = index
        this.$http.get('/api/mounts/' + this.mounts[index].path.slice(0, -1)).then(function (response) {
          this.mountConfig = response.data.result
          this.mountConfigModified = this.mountConfig
        }, function (err) {
          openNotification({
            title: 'Error',
            message: err.body.error,
            type: 'danger'
          })
          console.log(err.body.error)
        })
      },

      postMountConfig: function () {
        // construct payload
        var address = '/api/mounts/' + this.mounts[this.selectedIndex].path.slice(0, -1)
        var parsed = JSON.parse(this.mountConfigModified)
        var body = {
          'default_lease_ttl': parsed.default_lease_ttl.toString(),
          'max_lease_ttl': parsed.max_lease_ttl.toString()
        }
        var payload = {
          headers: {
            'X-CSRF-Token': this.csrf
          }
        }

        // tune mount
        this.$http.post(address, body, payload).then(function (response) {
          // alert success
          openNotification({
            title: 'Success',
            message: 'Mount tuned',
            type: 'success'
          })

          // double check config from server
          address = '/api/mounts/' + this.mounts[this.selectedIndex].path.slice(0, -1)
          this.$http.get(address).then(function (response) {
            // update page data accordingly
            this.mountConfig = response.data.result
            this.mountConfigModified = this.mountConfig
            this.mounts[this.selectedIndex].conf.default_lease_ttl = this.mountConfig.default_lease_ttl
            this.mounts[this.selectedIndex].conf.max_lease_ttl = this.mountConfig.max_lease_ttl
          }, function (err) {
            openNotification({
              title: 'Error',
              message: err.body.error,
              type: 'danger'
            })
            console.log(err.body.error)
          })
        }, function (err) {
          openNotification({
            title: 'Error',
            message: err.body.error,
            type: 'danger'
          })
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

  .fa-trash-o {
    color: red;
  }

  .fa-info {
    color: lightskyblue;
  }

  .tooltip {
    display: inherit;
  }
</style>
