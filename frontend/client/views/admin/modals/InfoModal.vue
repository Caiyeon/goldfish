<template>
  <modal :visible="visible" @close="close">
    <div class="box">
      <article class="media">
        <div class="media-content">
          <div class="content">
            <p>

              <article v-if="infoIsJSON" class="message is-primary">
                <div class="message-header">
                  {{ title }}:
                </div>
                <pre class="is-paddingless" v-highlightjs="info"><code class="javascript"></code></pre>
              </article>

              <article v-else>
                <strong>{{ title }}</strong>
                <br>{{ info }}
              </article>
            </p>
          </div>
        </div>
      </article>
    </div>
  </modal>
</template>

<script>
import { Modal } from 'vue-bulma-modal'

export default {
  components: {
    Modal
  },

  props: {
    visible: Boolean,
    title: String,
    info: String,
    infoIsJSON: Boolean
  },

  methods: {
    close () {
      this.$emit('close')
    }
  }
}
</script>
