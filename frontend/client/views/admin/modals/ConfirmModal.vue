<template>
  <modal :visible="visible" @close="close">
    <div class="box">
      <article class="media">
        <div class="media-content">
          <div class="content">
            <p>
              <strong>{{ title }}</strong>
              <br>
              <a class="button is-danger" @click="confirmed">
                <span>Delete</span>
                <span class="icon">
                  <i class="fa fa-times"></i>
                </span>
              </a>
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
    info: String
  },

  methods: {
    close () {
      this.$emit('close')
    },
    confirmed () {
      this.$emit('confirmed')
    }
  }
}
</script>
