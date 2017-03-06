<template>
  <div>
    <div class="tile is-ancestor">
      <div class="tile is-parent">
        <article class="tile is-child box">

          <!-- navigation box -->
          <div class="box is-clearfix">

            <div class="columns">
              <div class="column is-fullwidth">
                <p class="control has-addons">

                  <!-- up button -->
                  <a class="button is-medium is-primary is-paddingless is-marginless" @click="changePathUp()">
                    <span class="icon is-paddingless is-marginless">
                      <i class="fa fa-angle-up is-paddingless is-marginless"></i>
                    </span>
                  </a>

                  <!-- navigation input -->
                  <input class="input is-medium is-expanded" type="text"
                  placeholder="Enter the path of a secret or directory"
                  v-model="currentPath"
                  @keyup.enter="changePath(currentPath)">

                </p>
              </div>
            </div>

            <!-- manual insertion button: to be implemented later -->
            <!-- <a class="button is-primary is-outlined">
              <span class="icon is-small">
                <i class="fa fa-plus"></i>
              </span>
              <span>Insert Secret</span>
            </a> -->

            <!-- legend -->
            <a class="tag is-danger is-unselectable is-disabled is-pulled-right">Mount</a>
            <a class="tag is-primary is-unselectable is-disabled is-pulled-right">Path</a>
            <a class="tag is-info is-unselectable is-disabled is-pulled-right">Secret</a>
            <a class="tag is-success is-unselectable is-disabled is-pulled-right">Key</a>
          </div>

          <!-- data table -->
          <div class="table-responsive">
            <table class="table is-striped is-narrow">

              <!-- headers -->
              <thead>
                <tr>
                  <th>Type</th>
                  <th v-for="header in tableHeaders">{{ header }}</th>
                </tr>
              </thead>

              <!-- body -->
              <tbody>
                <tr v-for="(entry, index) in tableData">
                  <td width="68">
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

                <!-- new key value pair insertion row -->
                <tr
                  v-show="currentPathType === 'Secret'"
                  @keyup.enter="addKeyValue()"
                >
                  <td width="68">
                  </td>
                  <td>
                    <p class="control">
                    <input
                      class="input is-small"
                      type="text"
                      placeholder="Add a key"
                      v-model="newKey"
                      v-bind:class="[
                        newKey === '' ? '' : 'is-success',
                        newKeyExists ? 'is-danger' : '']"
                    >
                    </p>
                  </td>
                  <td>
                    <p class="control">
                    <input
                      class="input is-small"
                      type="text"
                      placeholder="Add a value"
                      v-model="newValue"
                      v-bind:class="[newValue === '' ? '' : 'is-success']"
                    >
                    </p>
                  </td>
                </tr>

              </tbody>

              <!-- footer only shows beyond a certain amount of data -->
              <tfoot v-show="tableData.length > 10">
                <tr>
                  <th>Type</th>
                  <th v-for="header in tableHeaders">{{ header }}</th>
                </tr>
              </tfoot>

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
        currentPath: 'data/',
        tableHeaders: [],
        tableData: [],
        newKey: '',
        newValue: ''
      }
    },

    mounted: function () {
      this.changePath(this.currentPath)
    },

    computed: {
      currentPathType: function () {
        if (this.currentPath === '' || this.currentPath === '/') {
          return 'Mount'
        }
        if (this.currentPath.slice(-1) === '/') {
          return 'Path'
        } else {
          return 'Secret'
        }
      },

      newKeyExists: function () {
        for (var i = 0; i < this.tableData.length; i++) {
          if (this.tableData[i].path === this.newKey) {
            return true
          }
        }
        return false
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
        this.$http.get('/api/mounts')
          .then((response) => {
            this.tableData = []
            this.tableHeaders = ['Mounts', 'Description', '']
            this.csrf = response.headers['x-csrf-token']
            let result = response.data.result

            var keys = Object.keys(result)
            for (var i = 0; i < keys.length; i++) {
              this.tableData.push({
                path: keys[i],
                type: result[keys[i]]['type'],
                desc: result[keys[i]]['description'],
                conf: result[keys[i]]['config']
              })
            }
          })
          .catch((error) => {
            openNotification({
              title: 'Error',
              message: error.body.error,
              type: 'danger'
            })
            console.log(error.body.error)
          })
      },

      changePath: function (path) {
        this.newKey = ''
        this.newValue = ''

        if (path === '' || path === '/') {
          this.currentPath = ''
          this.getMounts()
          return
        }

        this.$http.get('/api/secrets?path=' + path)
          .then((response) => {
            this.tableData = []
            this.currentPath = path
            let result = response.data.result

            if (path.slice(-1) === '/') {
              // listing subdirectories
              this.tableHeaders = ['Subpaths', 'Description', '']
              for (var i = 0; i < result.length; i++) {
                this.tableData.push({
                  path: result[i],
                  type: result[i].slice(-1) === '/' ? 'Path' : 'Secret'
                })
              }
            } else {
              // listing key value pairs
              this.tableHeaders = ['Key', 'Value', '']
              var keys = Object.keys(result)
              for (var j = 0; j < keys.length; j++) {
                this.tableData.push({
                  path: keys[j],
                  type: 'Key',
                  desc: result[keys[j]]
                })
              }
            }
          })

          .catch((error) => {
            openNotification({
              title: 'Error',
              message: error.body.error,
              type: 'danger'
            })
            console.log(error.body.error)
          })
      },

      changePathUp: function () {
        // cut the trailing slash off if it exists
        var noTrailingSlash = this.currentPath
        if (this.currentPath.slice(-1) === '/') {
          noTrailingSlash = this.currentPath.substring(0, this.currentPath.length - 1)
        }
        // remove up to the last slash if it exists
        this.currentPath = noTrailingSlash.substring(0, noTrailingSlash.lastIndexOf('/')) + '/'
        // fetch data again
        this.changePath(this.currentPath)
      },

      type: function (index) {
        switch (this.tableData[index].type) {
          case 'Secret':
            return { 'tag': true, 'is-info': true }
          case 'Path':
            return { 'tag': true, 'is-primary': true }
          case 'Key':
            return { 'tag': true, 'is-success': true }
          default:
            return { 'tag': true, 'is-danger': true }
        }
      },

      addKeyValue: function () {
        if (this.newKey === '' || this.newValue === '') {
          openNotification({
            title: 'Invalid',
            message: 'key and value must be non-empty',
            type: 'warning'
          })
          return
        }

        if (this.newKeyExists) {
          openNotification({
            title: 'Invalid',
            message: 'key already exists',
            type: 'warning'
          })
          return
        }

        openNotification({
          title: 'SoonTM',
          message: 'insertion not yet implemented',
          type: 'danger'
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
