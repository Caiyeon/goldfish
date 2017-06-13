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
                     v-model="currToken"
                     @keyup.enter="unwrapToken()"
              >
            </p>
            <p class="control">
            <a class="button is-primary"
            @click="unwrapToken()"
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
                               @keyup.enter="doneEdit(entry.key,index)"
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
                               @keyup.enter="doneEdit(entry.key,index)">
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
                        keyExists(newKey) ? 'is-danger' : '']"
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
                 <div class="control">
                  <input class="input" type="text"
                  placeholder="Wrap-ttl e.g. '5m'"
                  v-model="wrap_ttl"
                  :class="stringToSeconds(this.wrap_ttl) < 0 ? 'is-danger' : ''">
                   <p v-if="stringToSeconds(this.wrap_ttl) < 0" class="help is-danger">
                  TTL cannot be negative
                </p>
                <p v-if="stringToSeconds(this.wrap_ttl) > 0" class="help is-info">
                  {{ stringToSeconds(this.wrap_ttl) }} seconds
                </p>
                </div>
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
const querystring = require('querystring')
export default {
  data () {
    return {
      csrf: '',
      tableData: [],
      currToken: '',
      newKey: '',
      newValue: '',
      wrap_ttl: '300'
    }
  },

  mounted: function () {
    // fetch csrf token upon mounting
    this.$http.get('/api/wrapping')
    .then((response) => {
      this.csrf = response.headers['x-csrf-token']
    })
    .catch((error) => {
      this.$onError(error)
    })
  },

  methods: {

    // takes out "isClicked" field in tableData so content can be sent off
    packData: function () {
      var data = {}
      for (var i = 0; i < this.tableData.length; i++) {
        data[this.tableData[i].key] = this.tableData[i].value
      }
      return data
    },

    wrapData: function () {
      // do nothing if the table is empty
      if (this.tableData.length === 0) {
        return
      }

      this.$http.post('/api/wrapping/wrap', querystring.stringify({
        wrapttl: this.wrap_ttl,
        data: JSON.stringify(this.packData())
      }), {
        headers: {'X-CSRF-Token': this.csrf}
      })
      .then((response) => {
        this.$message({
          message: 'Wrapping token: ' + response.data.result,
          type: 'success',
          duration: 0,
          showCloseButton: true
        })
      })
      .catch((error) => {
        this.$onError(error)
      })
    },

    unwrapToken: function () {
      // do nothing if there is no input token string
      if (this.currToken === '') {
        return
      }
      this.$http.post('/api/wrapping/unwrap', querystring.stringify({
        wrappingToken: this.currToken
      }), {
        headers: {'X-CSRF-Token': this.csrf}
      })
      .then((response) => {
        this.tableData = []
        this.unpackData(response.data.result)
      })
      .catch((error) => {
        this.$onError(error)
      })
    },

    // Extracts the received data (a map) into tableData format with isClicked field
    unpackData: function (rawTable) {
      Object.keys(rawTable).map((index) => this.tableData.push({
        key: index,
        value: rawTable[index],
        isClicked: false
      }))
    },

    deleteItem: function (index) {
      this.tableData.splice(index, 1)
    },

    clicked: function (index) {
      this.tableData[index].isClicked = true
    },

    // Returns true if the new key already exists in the current table
    keyExists: function (key) {
      for (var i = 0; i < this.tableData.length; i++) {
        if (this.tableData[i].key === key) {
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
          message: 'Key cannot be empty',
          type: 'warning'
        })
        return
      }
      if (this.keyExists(this.newKey)) {
        this.$notify({
          title: 'Invalid',
          message: 'Key already exists',
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

    doneEdit: function (key, index) {
      // check key and value again
      if (this.tableData[index].key === '') {
        this.$notify({
          title: 'Invalid',
          message: 'Key already exists',
          type: 'warning'
        })
        return
      }
      if (this.keyExists(key)) {
        this.$notify({
          title: 'Invalid',
          message: 'Key already exists',
          type: 'warning'
        })
        return
      }
      this.tableData[index].isClicked = false
    },

    stringToSeconds: function (str) {
      if (str.includes('-')) {
        return -1
      }
      var totalSeconds = 0
      var days = str.match(/(\d+)\s*d/)
      var hours = str.match(/(\d+)\s*h/)
      var minutes = str.match(/(\d+)\s*m/)
      var seconds = str.match(/(\d+)$/) || str.match(/(\d+)\s*s/)
      if (days) { totalSeconds += parseInt(days[1]) * 86400 }
      if (hours) { totalSeconds += parseInt(hours[1]) * 3600 }
      if (minutes) { totalSeconds += parseInt(minutes[1]) * 60 }
      if (seconds) { totalSeconds += parseInt(seconds[1]) }
      return totalSeconds
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
