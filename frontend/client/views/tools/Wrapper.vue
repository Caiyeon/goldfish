<template>
  <div>
    <div class="tile is-ancestor">
      <div class="tile is-parent">
        <article class="tile is-child box">
          <label class="label">Token</label>
          <div class="field has-addons">
            <p class="control is-expanded">
              <input class="input" type="text"
                     placeholder="Paste token here to unwrap"
                     v-model=currToken
              >
            </p>
            <p class="control is-pulled-right">
            <a class="button is-primary"
            @click="unWrapToken()"
            :disabled="currToken === ''">
            <span>Unwrap</span>
            </a>
            </p>
          </div>
          <label class="label"> Data </label>
          <article class="tile is-child box">

            <!-- data table -->

            <div class="table-responsive">
              <table class="table is-striped is-narrow">

                <!-- header -->

                <thead>
                  <tr>
                    <th>Key</th>
                    <th>Value</th>
                    <th width="1"></th>
                  </tr>
                </thead>
                <!-- body -->
                <tbody>
                  <tr v-for="(entry, index) in tableData">
                    <!-- Editable key field -->
                    <td v-if="entry.isClicked">
                      <p class="control">
                        <input class="input is-small"
                               type="text" placeholder="" v-model="entry.key"
                               @keyup.enter="doneEdit(index)"
                        >
                      </p>
                    </td>

                    <!-- View-only -->
                    <td v-else @click="clicked(index)">
                      {{ entry.key }}
                    </td>

                    <!-- Editable value field -->
                    <td v-if="entry.isClicked">
                      <p class="control">
                        <input class="input is-small" type="text" placeholder="" v-model="entry.value"
                               @keyup.enter="doneEdit(index)">
                      </p>
                    </td>

                    <!-- View-only -->
                    <td v-else @click="clicked(index)">
                      {{ entry.value }}
                    </td>


                    <td width="30">
                      <a v-if="entry.isClicked" @click="deleteItem(index)">
                        <span class="icon">
                          <i class="fa fa-times-circle"></i>
                        </span>
                      </a>
                    </td>

                  </tr>

                  <!-- new key value pair insertion row -->
                  <tr @keyup.enter="addKeyValue()">
                    <td>
                      <p class="control">
                        <input
                        class="input is-small"
                        type="text"
                        placeholder="Add a key to wrap"
                        v-model="newKey"
                        v-bind:class="[
                        newKey === '' ? '' : 'is-success',
                        newKeyExists() ? 'is-danger' : '']"
                        >
                      </p>
                    </td>
                    <td>

                      <p class="control">
                        <input
                        class="input is-small"
                        type="text"
                        placeholder="Add a value to wrap"
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
                    <th>Key</th>
                    <th>Value</th>
                    <th></th>
                  </tr>
                </tfoot>
              </table>
            </div>

            <nav class="level">
              <div class="level-left"></div>
              <div class="level-right">
               <div class="field has-addons is-pulled-right">
                 <p class="control">
                  <input class="input" type="text"
                  placeholder="Wrap-ttl e.g. '5m'"
                  v-model=wrap_ttl
                  >
                </p>
                <p class="control">
                  <a class="button is-primary"
                  @click="wrapData()"
                  :disabled="tableData.length === 0">
                  <span>Wrap</span>
                </a>
              </p>
            </div>
          </div>
        </nav>
          </article>
        </article>
      </div>
    </div>
  </div>
</template>

<script>
// const querystring = require('querystring')
export default {
  data () {
    return {
      csrf: '',
      tableData: [],
      currToken: '',
      newKey: '',
      newValue: '',
      wrap_ttl: ''
    }
  },

  mounted: function () {
    this.$notify({
      title: 'Under Construction',
      message: 'This page doesn\'t work.',
      type: 'warning'
    })

    // fetch csrf token upon mounting
    // this.$http.get('/api/users/csrf')
    // .then((response) => {
    //   this.csrf = response.headers['x-csrf-token']
    // })
    // .catch((error) => {
    //   this.$onError(error)
    // })
  },

  methods: {
    wrapData: function () {
      // if insertion row (last row of table) is not empty, add to tableData before calling API to wrap
      if (this.newKey !== '' || this.newValue !== '') {
        // helper function will take care of edge cases
        this.addKeyValue()
      }

    //   this.$http.post('/api/wrapping/wrap', querystring.stringify({

    //   // wrapttl takes value of user's input
    //     wrapttl: this.wrap_ttl,
    //     data: {
    //       key: "value",
    //       anotherkey: "anothervalue"
    //       }
    //     }),
    //     {
    //     headers: {'X-CSRF-Token': this.csrf}
    //   })
    //   .then((response) => {
    //   // wrapping token:
    //   console.log(response.data.result)
    // })
    //   .catch((error) => {
    //     this.$onError(error)
    //   })
    },

    unWrapToken: function () {
      // body...
    },

    deleteItem: function (index) {
      this.tableData.splice(index, 1)
    },

    clicked: function (index) {
      this.tableData[index].isClicked = true
    },

    // Returns true if the new key already exists in the current table
    newKeyExists: function () {
      for (var i = 0; i < this.tableData.length; i++) {
        if (this.tableData[i].key === this.newKey) {
          return true
        }
      }
      return false
    },

    addKeyValue: function () {
      // check if key is not empty
      if (this.newKey === '') {
        this.$notify({
          title: 'Invalid',
          message: 'key cannot be empty',
          type: 'warning'
        })
        return
      }
      if (this.newKeyExists()) {
        this.$notify({
          title: 'Invalid',
          message: 'key already exists',
          type: 'warning'
        })
        return
      }

      // insert new key value pair to table data
      this.tableData.push({
        key: this.newKey,
        value: this.newValue,
        isClicked: false
      })
      // reset so that a new pair can be inserted
      this.newKey = ''
      this.newValue = ''
    },

    doneEdit: function (index) {
      // check key and value again
      if (this.tableData[index].key === '') {
        this.$notify({
          title: 'Invalid',
          message: 'Edits can\'t cause key to be empty',
          type: 'warning'
        })
        return
      }
      if (this.changesConflict(index)) {
        this.$notify({
          title: 'Invalid',
          message: 'Edits can\'t cause key to duplicate',
          type: 'warning'
        })
        return
      }
      this.tableData[index].isClicked = false
    },

    changesConflict: function (index) {
      for (var i = 0; i < this.tableData.length; i++) {
        if (i !== index && this.tableData[i].key === this.tableData[index].key) {
          return true
        }
      }
      return false
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
</style>
