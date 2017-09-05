<template>
  <div>
    <div class="tile is-ancestor">
      <div class="tile is-parent">
        <article class="tile is-child box">

          <!-- navigation box -->
          <div class="box is-clearfix">

            <div class="columns">
              <div class="column is-fullwidth">
                <div class="field has-addons">
                  <!-- up button -->
                  <p class="control">
                  <a class="button is-medium is-primary is-paddingless is-marginless" @click="changePathUp()">
                    <span class="icon is-paddingless is-marginless">
                      <i class="fa fa-angle-up is-paddingless is-marginless"></i>
                    </span>
                  </a>
                  </p>
                  <p class="control is-expanded">
                  <!-- navigation input -->
                  <input class="input is-medium is-expanded" type="text"
                  placeholder="Enter the path of a secret or directory"
                  v-model.lazy="currentPath"
                  @keyup.enter="changePath(currentPath)">
                  </p>
                </div>
              </div>
            </div>

            <!-- manual insertion button: to be implemented later -->
            <!-- <a class="button is-primary is-outlined">
              <span class="icon is-small">
                <i class="fa fa-plus"></i>
              </span>
              <span>Insert Secret</span>
            </a> -->

            <!-- Actions on current path -->
            <a v-if="editMode === false && currentPathType === 'Path'"
              class="button is-info is-small is-marginless"
              v-on:click="startEdit">
              Add Secret
            </a>

            <!-- Actions on current secret -->
            <a v-if="editMode === false && currentPathType === 'Secret'"
              class="button is-success is-small is-marginless"
              v-on:click="startEdit"
              :disabled="displayJSON">
              Edit Secret
            </a>
            <a v-if="editMode === false && currentPathType === 'Secret'"
              class="button is-danger is-small is-marginless"
              v-on:click="deleteSecret(currentPath)">
              Delete Secret
            </a>
            <a v-if="editMode === false && currentPathType === 'Secret'"
              class="button is-info is-small is-marginless"
              v-on:click="displayJSON = !displayJSON">
              Display JSON
            </a>

            <!-- Edit mode buttons -->
            <a v-if="editMode === true && currentPathType === 'Secret'"
              class="button is-success is-small is-marginless"
              v-on:click="saveEdit">
              Save Secret
            </a>
            <a v-if="editMode === true"
              class="button is-warning is-small is-marginless"
              v-on:click="cancelEdit">
              Cancel Edit
            </a>
          </div>

          <!-- data table -->
          <article v-if="!displayJSON">
            <table class="table is-fullwidth is-striped is-narrow">

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
                    <span class="tag is-rounded is-pulled-left" v-bind:class="type(index)">
                      {{ entry.type }}
                    </span>
                  </td>

                  <!-- Editable key field -->
                  <td v-if="editMode && currentPathType === 'Secret'">
                    <p class="control">
                      <input class="input is-small" type="text" placeholder="" v-model="entry.path">
                    </p>
                  </td>
                  <!-- View-only -->
                  <td v-else>
                    <span v-if="currentPathType === 'Secret'">
                      {{ entry.path }}
                    </span>
                    <a v-else @click="changePath(currentPath, entry)">
                      {{ entry.path }}
                    </a>
                  </td>

                  <!-- Editable value field -->
                  <td v-if="editMode && currentPathType === 'Secret'">
                    <p class="control">
                      <input class="input is-small" type="text" placeholder="" v-model="entry.desc">
                    </p>
                  </td>
                  <!-- View-only -->
                  <td v-if="!editMode && currentPathType === 'Secret'">
                    {{ entry.desc }}
                  </td>

                  <!-- Save some space for deletion button -->
                  <td width="68">
                    <!-- Deleting a key-value pair in edit mode -->
                    <a v-if="editMode && currentPathType === 'Secret'" @click="deleteItem(index)">
                    <span class="icon">
                      <i class="fa fa-times-circle"></i>
                    </span>
                    </a>

                    <!-- Deleting a secret -->
                    <a v-if="confirmDelete.includes(entry.path)"
                    @click="deleteSecret(currentPath + entry.path)">
                      <span class="tag is-rounded is-danger is-pulled-right">
                        Confirm
                      </span>
                    </a>

                    <a v-else-if="currentPathType === 'Path' && entry.type === 'Secret'"
                    @click="confirmDelete.push(entry.path)">
                      <span class="icon is-pulled-right">
                        <i class="fa fa-trash-o"></i>
                      </span>
                    </a>
                  </td>
                </tr>

                <!-- new key value pair insertion row -->
                <tr
                  v-if="editMode && currentPathType === 'Secret'"
                  @keyup.enter="addKeyValue()"
                >
                  <td width="68">
                  </td>
                  <td>
                    <p class="control">
                    <input v-focus
                      class="input is-small"
                      type="text"
                      ref="newKey"
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

                <!-- new secret insertion -->
                <tr
                  v-if="editMode && currentPathType === 'Path'"
                  @keyup.enter="addSecret()"
                >
                  <td width="68">
                  </td>
                  <td>
                    <p class="control">
                    <input v-focus
                      class="input is-small"
                      type="text"
                      placeholder="Press enter to add a secret"
                      v-model="newKey"
                      v-bind:class="[
                        newKey === '' ? '' : 'is-success',
                        newKeyExists ? 'is-danger' : '']"
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
          </article>

          <article v-if="displayJSON" class="message is-primary">
            <div class="message-header">
              JSON:
            </div>
            <pre v-highlightjs="JSON.stringify(constructedPayload, null, '    ')"><code class="javascript"></code></pre>
          </article>

        </article>
      </div>
    </div>
  </div>
</template>

<script>
const querystring = require('querystring')

export default {
  data () {
    return {
      currentPath: '',
      currentPathCopy: '',
      displayJSON: false,
      tableData: [],
      tableDataCopy: [],
      newKey: '',
      newValue: '',
      editMode: false,
      confirmDelete: []
    }
  },

  mounted: function () {
    this.changePath(this.currentPath)
  },

  computed: {
    session: function () {
      return this.$store.getters.session
    },

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

    // Returns true if the new key already exists in the current secret
    newKeyExists: function () {
      for (var i = 0; i < this.tableData.length; i++) {
        if (this.tableData[i].path === this.newKey) {
          return true
        }
      }
      return false
    },

    // Returns a constructed payload for writing secrets
    constructedPayload: function () {
      if (this.currentPathType === 'Secret') {
        var data = {}
        for (var i = 0; i < this.tableData.length; i++) {
          data[this.tableData[i].path] = this.tableData[i].desc
        }
        return data
      } else {
        return {}
      }
    },

    tableHeaders: function () {
      if (this.currentPathType === 'Secret') {
        return ['Key', 'Value', '']
      } else if (this.currentPathType === 'Path') {
        return ['Subpaths', '']
      }
      return []
    }
  },

  methods: {
    deleteItem: function (index) {
      this.tableData.splice(index, 1)
    },

    changePath: function (path, entry) {
      if (entry) {
        if (entry.type === 'Key') {
          return
        } else {
          path += entry.path
        }
      }

      this.newKey = ''
      this.newValue = ''
      this.editMode = false
      this.displayJSON = false
      this.confirmDelete = []

      this.$http.get('/v1/secrets?path=' + encodeURIComponent(path), {
        headers: {'X-Vault-Token': this.session ? this.session.token : ''}
      })
      .then((response) => {
        this.tableData = []
        this.currentPath = response.data.path
        let result = response.data.result
        if (this.currentPathType === 'Path') {
          // listing subdirectories
          for (var i = 0; i < result.length; i++) {
            this.tableData.push({
              path: result[i],
              type: result[i].slice(-1) === '/' ? 'Path' : 'Secret'
            })
          }
        } else {
          // listing key value pairs
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
        this.$onError(error)
        this.tableData = []
      })
    },

    changePathUp: function () {
      // cut the trailing slash off if it exists
      let noTrailingSlash = this.currentPath
      if (this.currentPath.slice(-1) === '/') {
        noTrailingSlash = this.currentPath.substring(0, this.currentPath.length - 1)
      }

      // remove up to the last slash if it exists
      let resultPath = noTrailingSlash.substring(0, noTrailingSlash.lastIndexOf('/')) + '/'
      if (resultPath === '/') {
        this.$notify({
          title: 'Invalid',
          message: 'Already at top-level of mount',
          type: 'warning'
        })
        return
      }

      // fetch data again
      this.currentPath = resultPath
      this.changePath(this.currentPath)
    },

    type: function (index) {
      switch (this.tableData[index].type) {
        case 'Secret':
          return { 'tag': true, 'is-rounded': true, 'is-info': true }
        case 'Path':
          return { 'tag': true, 'is-rounded': true, 'is-primary': true }
        case 'Key':
          return { 'tag': true, 'is-rounded': true, 'is-success': true }
        default:
          return { 'tag': true, 'is-rounded': true, 'is-danger': true }
      }
    },

    addKeyValue: function () {
      // only allow insertion if key and value are valid
      if (this.newKey === '') {
        this.$notify({
          title: 'Invalid',
          message: 'Key is required',
          type: 'warning'
        })
        return
      }
      if (this.newKeyExists) {
        this.$notify({
          title: 'Invalid',
          message: 'Key already exists',
          type: 'warning'
        })
        return
      }
      // insert new key value pair to local table (don't write it to server yet)
      this.tableData.push({
        path: this.newKey,
        type: 'Key',
        desc: this.newValue
      })
      // reset so that a new pair can be inserted
      this.newKey = ''
      this.newValue = ''
      // reset focus to key input
      this.$refs.newKey.focus()
    },

    startEdit: function () {
      if (this.displayJSON) {
        return
      }
      this.editMode = true
      this.currentPathCopy = this.currentPath
      // a deep copy is needed in case the edit is cancelled
      this.tableDataCopy = JSON.parse(JSON.stringify(this.tableData))
    },

    saveEdit: function () {
      // if there is a current new key/value pair, add it in first
      if (!(this.newKey === '' || this.newValue === '') && !this.newKeyExists) {
        this.addKeyValue()
      }
      var body = JSON.stringify(this.constructedPayload)
      this.$http.post('/v1/secrets?path=' + encodeURIComponent(this.currentPath), querystring.stringify({
        body: body
      }), {
        headers: {'X-Vault-Token': this.session ? this.session.token : ''}
      })
      .then((response) => {
        this.$notify({
          title: 'Success!',
          message: '',
          type: 'success'
        })
        this.editMode = false
      })
      .catch((error) => {
        this.$onError(error)
      })
    },

    cancelEdit: function () {
      this.editMode = false
      this.tableData = this.tableDataCopy
      this.currentPath = this.currentPathCopy
    },

    addSecret: function () {
      // only allow insertion if key and value are valid
      if (this.newKey === '') {
        this.$notify({
          title: 'Invalid',
          message: 'key and value must be non-empty',
          type: 'warning'
        })
        return
      }

      // Backup in case edit is cancelled
      this.currentPathCopy = this.currentPath
      this.tableDataCopy = JSON.parse(JSON.stringify(this.tableData))

      // Display the to-be path of the new secret
      this.currentPath += this.newKey
      this.newKey = ''

      // Give the user a proper secret editing UI
      this.editMode = true
      this.tableData = []

      // Warn the user that this secret is all a draft until saved
      this.$notify({
        title: 'This is a draft!',
        message: 'Click save secret to finalize',
        type: 'warning',
        duration: 10000
      })
    },

    deleteSecret: function (path) {
      // check if current path is valid
      if (!path.includes('/')) {
        this.$notify({
          title: 'Invalid',
          message: 'Cannot delete a mount',
          type: 'warning'
        })
        return
      }

      // recursive deletion may come later, but not now
      if (path.endsWith('/')) {
        this.$notify({
          title: 'Invalid',
          message: 'Cannot delete a path',
          type: 'warning'
        })
        return
      }

      // request deletion of secret
      this.$http.delete('/v1/secrets?path=' + encodeURIComponent(path), {
        headers: {'X-Vault-Token': this.session ? this.session.token : ''}
      })
      .then((response) => {
        this.$notify({
          title: 'Success!',
          message: 'Secret deleted!',
          type: 'success'
        })
        this.editMode = false

        if (this.currentPath === path) {
          // if deleting current secret, wipe table data
          this.tableData = []
        } else {
          // if deleting a row, find it and remove it
          for (var i = 0; i < this.tableData.length; i++) {
            if (this.currentPath + this.tableData[i].path === path) {
              this.deleteItem(i)
            }
          }
        }
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

  .fa-times-circle {
    color: red;
  }
</style>
