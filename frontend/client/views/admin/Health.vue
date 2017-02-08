<template>
  <div>
    <div class="tile is-ancestor">

      <div class="tile is-parent">
        <article class="tile is-child box">
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
                  <tr v-for="key in tableKeys">
                    <td>
                      {{ key }}
                    </td>
                    <td>
                      {{ tableData[key] }}
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
        path: '/generic/path',
        tableData: {
          'cluster_id': 'c9abceea-4f46-4dab-a688-5ce55f89e228',
          'cluster_name': 'vault-cluster-5515c810',
          'version': '0.6.1-dev',
          'server_time_utc': 1469555798,
          'standby': false,
          'sealed': false,
          'initialized': true
        }
      }
    },
    mounted: function () {
      openNotification({
        title: 'Under construction',
        message: 'Page is not implemented yet',
        type: 'danger'
      })
    },
    computed: {
      tableKeys: function () {
        return Object.keys(this.tableData)
      }
    },
    methods: {
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
