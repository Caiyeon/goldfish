<template>
  <div>
    <div class="tile is-ancestor">

      <div class="tile is-parent">
        <article class="tile is-child box">

          <div class="control has-icon is-horizontal">
            <span class="icon is-medium">
              <i class="fa fa-angle-up"></i>
            </span>
            <h3 class="title is-3">
              {{ currentPath }}
            </h3>
          </div>

          <div class="box">
            <div class="table-responsive">
              <table class="table is-striped is-narrow">
                <thead>
                  <tr>
                    <th>Directory</th>
                    <th>Description</th>
                    <th></th>
                  </tr>
                </thead>
                <tbody>
                  <tr v-for="(entry, index) in tableData">
                    <td>
                      <a @click="changePath(entry.path)">
                        {{ entry.path }}
                      </a>
                    </td>
                    <td>
                      {{ entry.desc }}
                    </td>
                    <td class="is-icon">
                      <a @click="deleteItem">
                        <i class="fa fa-trash-o"></i>
                      </a>
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
        currentPath: '',
        tableData: []
      }
    },

    mounted: function () {
      openNotification({
        title: 'Under construction',
        message: 'Page is not implemented yet',
        type: 'danger'
      })

      this.$http.get('/api/mounts').then(function (response) {
        this.tableData = []
        this.csrf = response.headers.get('x-csrf-token')

        var keys = Object.keys(response.data.result)
        for (var i = 0; i < keys.length; i++) {
          this.tableData.push({
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

    computed: {
      tableKeys: function (index) {
        return Object.keys(this.tableData)[index]
      }
    },

    methods: {
      deleteItem: function () {
        openNotification({
          title: 'Under construction',
          message: 'Deletion is not supported yet',
          type: 'danger'
        })
      },

      changePath: function (path) {
        this.$http.get('/api/secrets?path=' + this.currentPath + path).then(function (response) {
          this.tableData = []
          this.currentPath = this.currentPath + path

          if (path.slice(-1) === '/') {
            // listing subdirectories
            for (var i = 0; i < response.data.result.length; i++) {
              this.tableData.push({
                path: response.data.result[i],
                desc: response.data.result[i].slice(-1) === '/' ? '' : 'secret'
              })
            }
          } else {
            // listing key value pairs
            var keys = Object.keys(response.data.result)
            for (var j = 0; j < keys.length; j++) {
              this.tableData.push({
                path: keys[j],
                desc: response.data.result[keys[j]]
              })
            }
          }
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
</style>
