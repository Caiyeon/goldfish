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
                  @keyup.enter="pushPath(currentPath)">
                  </p>
                </div>
              </div>
            </div>

            <!-- Actions on current path -->
            <a v-if="editMode === false && currentPathType === 'Path'"
              class="button is-info is-small is-marginless"
              v-on:click="startEdit()">
              Add Secret
            </a>
            <a v-if="editMode === false && currentPathType === 'Path'"
              class="button is-info is-small is-marginless"
              v-on:click="selectAllSecrets()">
              Select All Secrets
            </a>
            <a v-if="editMode === false && currentPathType === 'Path' && selectedRows.length !== 0"
              class="button is-warning is-small is-marginless"
              v-on:click="selectedRows = []; confirmDeleteSecrets = false">
              Cancel Selection
            </a>
            <a v-if="editMode === false && currentPathType === 'Path' && selectedRows.length !== 0 && confirmDeleteSecrets === false"
              class="button is-danger is-small is-marginless"
              v-on:click="confirmDeleteSecrets = true">
              Delete Selection
            </a>
            <a v-if="editMode === false && currentPathType === 'Path' && selectedRows.length !== 0 && confirmDeleteSecrets === true"
              class="button is-danger is-small is-marginless"
              v-on:click="deleteSelection()">
              Really Delete {{selectedRows.length}} Secrets?
            </a>

            <!-- Actions on current secret -->
            <a v-if="editMode === false && currentPathType === 'Secret'"
              class="button is-success is-small is-marginless"
              v-on:click="startEdit()"
              :disabled="displayJSON">
              Edit Secret
            </a>
            <a v-if="editMode === false && currentPathType === 'Secret' && confirmDeleteSecrets === false"
              class="button is-danger is-small is-marginless"
              v-on:click="confirmDeleteSecrets = true">
              Delete Secret
            </a>
            <a v-if="editMode === false && currentPathType === 'Secret' && confirmDeleteSecrets === true"
              class="button is-danger is-small is-marginless"
              v-on:click="deleteSecret(currentPath)">
              Confirm Deletion
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

            <p v-if="editMode && currentPathType === 'Secret'" class="help is-info">
              Inputs are multi-line by default. Press tab to complete a key-value pair.
            </p>
          </div>

          <!-- data table -->
          <article v-if="!displayJSON">
            <table class="table is-fullwidth is-striped is-narrow">

              <!-- headers -->
              <thead>
                <tr>
                  <th @click="sortBy('type')">
                    Type
                    <i v-if="sortKey.key === 'type' && sortKey.order === 'asc'" class="fa fa-caret-up"></i>
                    <i v-if="sortKey.key === 'type' && sortKey.order === 'desc'" class="fa fa-caret-down"></i>
                  </th>
                  <th @click="sortBy('path')">
                    {{currentPathType === 'Secret' ? 'Key' : 'Subpaths'}}
                    <i v-if="sortKey.key === 'path' && sortKey.order === 'asc'" class="fa fa-caret-up"></i>
                    <i v-if="sortKey.key === 'path' && sortKey.order === 'desc'" class="fa fa-caret-down"></i>
                  </th>
                  <th v-if="currentPathType === 'Secret'" @click="sortBy('desc')">
                    Value
                    <i v-if="sortKey.key === 'desc' && sortKey.order === 'asc'" class="fa fa-caret-up"></i>
                    <i v-if="sortKey.key === 'desc' && sortKey.order === 'desc'" class="fa fa-caret-down"></i>
                  </th>
                  <th></th>
                </tr>
              </thead>

              <!-- body -->
              <tbody>
                <tr v-for="(entry, index) in sortedTableData"
                :class="selectedRows.includes(entry.path) ? 'is-selected' : ''">
                  <td width="68">
                    <span class="tag is-rounded is-pulled-left" v-bind:class="type(index)">
                      {{ entry.type }}
                    </span>
                  </td>

                  <!-- Editable key field -->
                  <td v-if="editMode && currentPathType === 'Secret'">
                    <p class="control">
                      <textarea style="font-family: monospace; padding: 3.5px 6.5px 3.5px 6.5px;"
                        v-bind:rows="String(entry.path).split('\n').length"
                        placeholder="" v-model="entry.path"
                        class="textarea is-small" type="text">
                      </textarea>
                    </p>
                  </td>
                  <!-- View-only -->
                  <td v-else @click="select(entry.path)">
                    <span
                      v-if="currentPathType === 'Secret'"
                      style="font-family: monospace;"
                    >
                      {{ entry.path }}
                    </span>
                    <a
                      v-else
                      @click="pushPath(currentPath + entry.path); select(entry.path)"
                      style="font-family: monospace;"
                    >
                      {{ entry.path }}
                    </a>
                  </td>

                  <!-- Editable value field -->
                  <td v-if="editMode && currentPathType === 'Secret'">
                    <p class="control">
                      <textarea style="font-family: monospace; padding: 3.5px 6.5px 3.5px 6.5px;"
                        v-focus
                        v-bind:rows="String(entry.desc).split('\n').length"
                        class="textarea is-small" type="text" placeholder="" v-model="entry.desc">
                      </textarea>
                    </p>
                  </td>
                  <!-- View-only -->
                  <td v-if="!editMode && currentPathType === 'Secret'"
                    style="white-space: pre-wrap; font-family: monospace;"
                    >{{ entry.desc }}</td>

                  <!-- Save some space for deletion button -->
                  <td width="68">
                    <!-- Deleting a key-value pair in edit mode -->
                    <a v-if="editMode && currentPathType === 'Secret'"
                      @click="deleteKeyPair(entry)">
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
                >
                  <td width="68">
                  </td>
                  <td>
                    <p class="control">
                    <textarea style="font-family: monospace; padding: 3.5px 6.5px 3.5px 6.5px;"
                      class="textarea is-small"
                      type="text"
                      ref="newKeyField"
                      placeholder="Add a key"
                      v-bind:rows="String(newKey).split('\n').length"
                      v-model="newKey"
                      v-bind:class="[
                        newKey === '' ? '' : 'is-success',
                        newKeyExists ? 'is-danger' : '']">
                    </textarea>
                    </p>
                  </td>
                  <td>
                    <p class="control">
                    <textarea style="font-family: monospace; padding: 3.5px 6.5px 3.5px 6.5px;"
                      class="textarea is-small"
                      type="text"
                      placeholder="Add a value"
                      ref="newValueField"
                      v-bind:rows="String(newValue).split('\n').length"
                      v-model="newValue"
                      v-on:keydown.tab.exact.prevent="addKeyValue()"
                      v-bind:class="[newValue === '' ? '' : 'is-success']">
                    </textarea>
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
                <tr v-if="this.currentPathType === 'Secret'">
                  <th>Type</th>
                  <th>Key</th>
                  <th>Value</th>
                  <th></th>
                </tr>
                <tr v-if="this.currentPathType === 'Path'">
                  <th>Type</th>
                  <th>Subpaths</th>
                  <th></th>
                </tr>
              </tfoot>

            </table>
          </article>

          <article v-if="displayJSON" class="message is-primary">
            <div class="message-header">
              JSON:
            </div>
            <pre class="is-paddingless" v-highlightjs="JSON.stringify(constructedPayload, null, '    ')"><code class="javascript"></code></pre>
          </article>

        </article>
      </div>
    </div>
  </div>
</template>

<script>
const querystring = require('querystring')
const _ = require('lodash')

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
      confirmDelete: [],
      confirmDeleteSecrets: false,
      selectedRows: [],
      sortKey: {
        key: '',
        order: ''
      }
    }
  },

  mounted: function () {
    // if path parameter was provided via url, go to that
    this.changePath(this.$route.query['path'] || this.currentPath)
  },

  watch: {
    // watch for route changes, e.g. if query parameters are updated
    '$route' (to, from) {
      // if query path is provided, go to that secret
      this.changePath(to.query['path'] || '')
    }
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

    sortedTableData: function () {
      if (!this.tableData || this.tableData.length === 0 || this.sortKey.key === '') {
        return this.tableData
      }
      return _.orderBy(this.tableData, [this.sortKey.key], [this.sortKey.order])
    }
  },

  methods: {
    tempFinish: function () {
      this.$nprogress.done()
    },

    deleteItem: function (index) {
      this.tableData.splice(index, 1)
    },

    deleteKeyPair: function (entry) {
      let index = _.findIndex(this.tableData, entry)
      if (index !== -1) {
        this.deleteItem(index)
      }
    },

    pushPath: function (path) {
      if (path) {
        this.$router.push({
          query: {
            path: path
          }
        })
      }
    },

    changePath: function (path) {
      // if user was editing, cancel it and restore local data
      if (this.editMode) {
        this.cancelEdit()
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
        this.selectedRows = []
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

      // update query parameter which will trigger loading the secret
      this.pushPath(resultPath)
    },

    type: function (index) {
      switch (this.sortedTableData[index].type) {
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
        this.$refs.newKeyField.focus()
        return
      }
      if (this.newKeyExists) {
        this.$notify({
          title: 'Invalid',
          message: 'Key already exists',
          type: 'warning'
        })
        this.$refs.newKeyField.focus()
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
      // using nextTick because vuejs mount order comes after focus
      this.$nextTick(() => this.$refs.newKeyField.focus())
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
        this.pushPath(this.currentPath)
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
    },

    deleteSecretMulti: function (paths) {
      var successes = 0
      var failures = 0

      for (var i = 0; i < paths.length; i++) {
        let path = paths[i]

        // check if current path is valid
        if (!path.includes('/')) {
          this.$notify({
            title: 'Invalid',
            message: 'Cannot delete a mount',
            type: 'warning'
          })
          failures++
          continue
        }
        // recursive deletion may come later, but not now
        if (path.endsWith('/')) {
          this.$notify({
            title: 'Invalid',
            message: 'Cannot delete a path',
            type: 'warning'
          })
          failures++
          continue
        }

        // request deletion of secret
        this.$http.delete('/v1/secrets?path=' + encodeURIComponent(path), {
          headers: {'X-Vault-Token': this.session ? this.session.token : ''}
        })
        .then((response) => {
          this.editMode = false
          successes++

          // if all requests have been completed, notify user
          if (successes + failures === paths.length) {
            this.$message({
              message: successes.toString() + ' out of ' + paths.length.toString() + ' secret(s) deleted successfully!',
              type: successes === paths.length ? 'success' : 'warning',
              duration: 0,
              showCloseButton: true
            })
          }

          if (this.currentPath === path) {
            // if deleting current secret, wipe table data
            this.tableData = []
          } else {
            // if deleting a row, find it and remove it
            for (var j = 0; j < this.tableData.length; j++) {
              if (this.currentPath + this.tableData[j].path === path) {
                this.deleteItem(j)
              }
            }
          }
        })
        .catch((error) => {
          this.$onError(error)
          failures++

          // if all requests have been completed, notify user
          if (successes + failures === paths.length) {
            this.$message({
              message: successes.toString() + ' out of ' + paths.length.toString() + ' secret(s) deleted successfully!',
              type: successes === paths.length ? 'success' : 'warning',
              duration: 0,
              showCloseButton: true
            })
          }
        })
      }
    },

    // either adds or removes the entry from selected rows
    select: function (entry) {
      // selection on keys is pointless
      if (this.currentPathType === 'Secret') {
        return
      }
      // selection on paths is not supported for now
      if (entry.endsWith('/')) {
        return
      }
      // reset delete confirmation
      this.confirmDeleteSecrets = false
      // otherwise, select the entry (or unselect it if it already is selected)
      if (this.selectedRows.includes(entry)) {
        this.unselect(entry)
      } else {
        this.selectedRows.push(entry)
      }
    },

    unselect: function (entry) {
      let index = this.selectedRows.indexOf(entry)
      if (index > -1) {
        this.selectedRows.splice(index, 1)
      }
    },

    selectAllSecrets: function () {
      // if this current path is not a path, there's nothing to be selected
      if (this.currentPathType !== 'Path') {
        return
      }
      // for each item in table, if it's a secret, add it to the selected array
      for (var i = 0; i < this.tableData.length; i++) {
        let entry = this.tableData[i].path
        if (!entry.endsWith('/') && !this.selectedRows.includes(entry)) {
          this.selectedRows.push(entry)
        }
      }
    },

    deleteSelection: function () {
      if (this.currentPathType !== 'Path') {
        return
      }

      // append full secret paths to an array
      var paths = []
      for (var i = 0; i < this.selectedRows.length; i++) {
        paths.push(this.currentPath + this.selectedRows[i])
      }

      // delete all selected secrets and give toast notification
      this.deleteSecretMulti(paths)

      // reset selection
      this.selectedRows = []
    },

    sortBy: function (s) {
      if (s === '') {
        this.sortKey = {
          key: '',
          order: ''
        }
        return
      }

      if (s === this.sortKey.key) {
        if (this.sortKey.order === '') {
          this.sortKey.order = 'asc'
        } else if (this.sortKey.order === 'asc') {
          this.sortKey.order = 'desc'
        } else {
          // the third sort click should reset sorting
          this.sortKey = {
            key: '',
            order: ''
          }
        }
      } else {
        this.sortKey = {
          key: s,
          order: 'asc'
        }
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

  .fa-trash-o {
    color: red;
  }

  .fa-times-circle {
    color: red;
  }
</style>
