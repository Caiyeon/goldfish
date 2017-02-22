<template>
  <div>
    <div class="tile is-ancestor">
      <div class="tile is-parent">
        <article class="tile is-child box">

          <div class="box is-clearfix">

            <div class="columns">
              <div class="column is-fullw">
                <p class="control has-addons">

                  <a class="button is-medium is-primary is-paddingless is-marginless" @click="changePathUp()">
                    <span class="icon is-paddingless is-marginless">
                      <i class="fa fa-angle-up is-paddingless is-marginless"></i>
                    </span>
                  </a>

                  <input class="input is-medium is-expanded" type="text"
                  placeholder="Enter the path of a secret or directory"
                  v-model="currentPath"
                  @keyup.enter="changePath(currentPath)">

                </p>
              </div>
            </div>

            <a class="tag is-danger is-unselectable is-disabled is-pulled-right">Mount</a>
            <a class="tag is-primary is-unselectable is-disabled is-pulled-right">Subdirectory</a>
            <a class="tag is-warning is-unselectable is-disabled is-pulled-right">Secret</a>

          </div>

          <div class="table-responsive">
            <table class="table is-striped is-narrow">
              <thead>
                <tr>
                  <th>Type</th>
                  <th v-for="header in tableHeaders">{{ header }}</th>
                </tr>
              </thead>
              <tbody>
                <tr v-for="(entry, index) in tableData">
                  <td class="is-icon">
                    <a class="tag is-disabled is-pulled-left" v-bind:class="type(index)">
                      {{ entry.type }}
                    </a>
                  </td>
                  <td>
                    <a v-bind:class="entry.type === 'Key' ? 'is-disabled' : ''" @click="changePath(currentPath + entry.path)">
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
        tableHeaders: [],
        tableData: []
      }
    },

    mounted: function () {
      this.getMounts()
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

      getMounts: function () {
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
          this.tableHeaders = ['Mounts', 'Description', '']
        }, function (err) {
          openNotification({
            title: 'Error',
            message: err.body.error,
            type: 'danger'
          })
          console.log(err.body.error)
        })
      },

      changePath: function (path) {
        if (path === '' || path === '/') {
          this.currentPath = ''
          this.getMounts()
          return
        }

        this.$http.get('/api/secrets?path=' + path).then(function (response) {
          this.tableData = []
          this.currentPath = path
          if (path.slice(-1) === '/') {
            // listing subdirectories
            for (var i = 0; i < response.data.result.length; i++) {
              this.tableData.push({
                path: response.data.result[i],
                type: response.data.result[i].slice(-1) === '/' ? 'Path' : 'Secret'
              })
            }
            this.tableHeaders = ['Subpaths', 'Description', '']
          } else {
            // listing key value pairs
            var keys = Object.keys(response.data.result)
            for (var j = 0; j < keys.length; j++) {
              this.tableData.push({
                path: keys[j],
                type: 'Key',
                desc: response.data.result[keys[j]]
              })
            }
            this.tableHeaders = ['Key', 'Value', '']
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

      changePathUp: function () {
        var noTrailingSlash = this.currentPath
        if (this.currentPath.slice(-1) === '/') {
          noTrailingSlash = this.currentPath.substring(0, this.currentPath.length - 1)
        }
        this.currentPath = noTrailingSlash.substring(0, noTrailingSlash.lastIndexOf('/')) + '/'
        this.changePath(this.currentPath)
      },

      type: function (index) {
        switch (this.tableData[index].type) {
          case 'Path':
            return { 'tag': true, 'is-primary': true }
          case 'Secret':
            return { 'tag': true, 'is-warning': true }
          case 'Key':
            return { 'tag': true, 'is-success': true }
          default:
            return { 'tag': true, 'is-danger': true }
        }
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
