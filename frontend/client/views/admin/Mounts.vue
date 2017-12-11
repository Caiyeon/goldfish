<template>
  <div>
    <div class="tile is-ancestor">

      <div class="tile is-parent is-vertical is-6">
        <article class="tile is-child box">
          <table class="table is-fullwidth is-striped is-narrow">
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
                <td width="68">
                  <span class="tag is-danger is-pulled-left">
                    {{ mount.type }}
                  </span>
                </td>
                <td>
                  <tooltip v-bind:label="mount.desc" placement="right" type="info" :rounded="true" >
                    <a @click="getMountConfig(index)">
                      {{ mount.path }}
                    </a>
                  </tooltip>
                </td>
                <td width="68">
                  <span class="tag is-primary is-pulled-left">
                    {{ mount.conf.default_lease_ttl === 0 ? 'Default' : mount.conf.default_lease_ttl }}
                  </span>
                </td>
                <td width="68">
                  <span class="tag is-primary is-pulled-left">
                    {{ mount.conf.max_lease_ttl === 0 ? 'Default' : mount.conf.max_lease_ttl }}
                  </span>
                </td>
              </tr>
            </tbody>
          </table>
        </article>
      </div>

      <div class="tile is-parent is-vertical is-6">
        <article class="tile is-child box">
          <h4 class="subtitle is-4">Mount Config</h4>

          <div class="field">
            <p class="control">
              <textarea class="textarea"
              placeholder="Select a mount"
              v-model="mountConfigModified"
              rows="5"></textarea>
            </p>
          </div>

          <div class="field">
            <p class="control is-pulled-right">
              <button @click="postMountConfig"
                class="button is-primary is-outlined"
                :disabled="mountConfig === mountConfigModified">
                <span>Submit Changes</span>
                <span class="icon is-small">
                  <i class="fa fa-check"></i>
                </span>
              </button>
            </p>
          </div>

        </article>
      </div>

    </div>

  </div>
</template>

<script>
import Tooltip from '../vue_bulma_modules/vue-bulma-tooltip'

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
    session: function () {
      return this.$store.getters.session
    }
  },

  mounted: function () {
    this.$http.get('/v1/mount', {
      headers: {'X-Vault-Token': this.session ? this.session.token : ''}
    })
    .then((response) => {
      this.mounts = []
      this.csrf = response.headers['x-csrf-token']
      let result = response.data.result
      let keys = Object.keys(result)
      for (var i = 0; i < keys.length; i++) {
        this.mounts.push({
          path: keys[i],
          type: result[keys[i]]['type'],
          desc: result[keys[i]]['description'],
          conf: result[keys[i]]['config']
        })
      }
    })
    .catch((error) => {
      this.$onError(error)
    })
  },

  methods: {
    getMountConfig: function (index) {
      this.selectedIndex = index
      this.$http.get('/v1/mount?mount=' + encodeURIComponent(this.mounts[index].path.slice(0, -1)), {
        headers: {'X-Vault-Token': this.session ? this.session.token : ''}
      })
      .then((response) => {
        this.mountConfig = JSON.stringify(response.data.result, null, 4)
        this.mountConfigModified = this.mountConfig
      })
      .catch((error) => {
        this.$onError(error)
      })
    },

    postMountConfig: function () {
      var address = '/v1/mount?mount=' + encodeURIComponent(this.mounts[this.selectedIndex].path.slice(0, -1))
      try {
        var parsed = JSON.parse(this.mountConfigModified)
      } catch (e) {
        this.$notify({
          title: 'Invalid',
          message: 'Could not parse JSON',
          type: 'warning'
        })
        return
      }

      this.$http.post(address, {
        default_lease_ttl: parsed.default_lease_ttl.toString(),
        max_lease_ttl: parsed.max_lease_ttl.toString()
      }, {
        headers: {
          'X-CSRF-Token': this.csrf,
          'X-Vault-Token': this.session ? this.session.token : ''
        }
      })

      .then((response) => {
        this.$notify({
          title: 'Success',
          message: 'Mount tuned',
          type: 'success'
        })
        // update page data accordingly
        this.$http.get(address, {
          headers: {'X-Vault-Token': this.session ? this.session.token : ''}
        }).then((response) => {
          this.mounts[this.selectedIndex].conf = response.data.result
          this.mountConfig = JSON.stringify(response.data.result, null, 4)
          this.mountConfigModified = this.mountConfig
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
