<template>
  <transition
    :name="animation"
    mode="out-in"
    appear
    :appear-active-class="transition.enterClass"
    :enter-active-class="transition.enterClass"
    :leave-active-class="transition.leaveClass"
  >
    <div
      role="tabpanel"
      v-show="realSelected"
      :class="classObject"
      :aria-labelledby="label"
      :aria-hidden="hidden"
    >
      <slot></slot>
    </div>
  </transition>
</template>

<script>
const TS = {
  'fade': {
    enterClass: 'fadeIn',
    leaveClass: 'fadeOut'
  },
  'fade-horizontal-rtl': {
    enterClass: 'fadeInRight',
    leaveClass: 'fadeOutLeft'
  },
  'fade-horizontal-ltr': {
    enterClass: 'fadeInLeft',
    leaveClass: 'fadeOutRight'
  },
  'slide-horizontal-rtl': {
    enterClass: 'slideInRight',
    leaveClass: 'slideOutLeft'
  },
  'slide-horizontal-ltr': {
    enterClass: 'slideInLeft',
    leaveClass: 'slideOutRight'
  },
  'fade-vertical-dtu': {
    enterClass: 'fadeInUp',
    leaveClass: 'fadeOutUp'
  },
  'fade-vertical-utd': {
    enterClass: 'fadeInDown',
    leaveClass: 'fadeOutDown'
  },
  'slide-vertical-dtu': {
    enterClass: 'slideInUp',
    leaveClass: 'slideOutUp'
  },
  'slide-vertical-utd': {
    enterClass: 'slideInDown',
    leaveClass: 'slideOutDown'
  }
}

export default {
  props: {
    icon: String,
    selected: Boolean,
    disabled: Boolean,
    label: {
      type: String,
      required: true
    },
    mode: {
      type: String,
      default: 'out-in'
    }
  },

  data () {
    return {
      realSelected: this.selected,
      transition: TS['fade']
    }
  },

  created () {
    this.$parent.tabPanes.push(this)
  },

  beforeDestroy () {
    this.$parent.tabPanes.splice(this.index, 1)
  },

  computed: {
    classObject () {
      const { realSelected } = this
      return {
        'tab-pane': true,
        'animated': true,
        'is-active': realSelected,
        'is-deactive': !realSelected
      }
    },
    layout () {
      return this.$parent.layout
    },
    direction () {
      if (this.layout === 'top' || this.layout === 'bottom') {
        return 'horizontal'
      } else if (this.layout === 'left' || this.layout === 'right') {
        return 'vertical'
      }
      return ''
    },
    animation () {
      return this.$parent.animation
    },
    onlyFade () {
      return this.$parent.onlyFade
    },
    gte () {
      if (this.direction === 'vertical') {
        return 'utd'
      } else if (this.direction === 'horizontal') {
        return 'ltr'
      }
      return ''
    },
    lt () {
      if (this.direction === 'vertical') {
        return 'dtu'
      } else if (this.direction === 'horizontal') {
        return 'rtl'
      }
      return ''
    },
    hidden () {
      return this.realSelected ? 'false' : 'true'
    },
    index () {
      return this.$parent.tabPanes.indexOf(this)
    }
  },

  watch: {
    '$parent.realSelectedIndex' (index, prevIndex) {
      if (!this.onlyFade) {
        let transition
        if (prevIndex === -1 || prevIndex > index) {
          transition = `${this.animation}${this.layout ? `-${this.direction}` : ''}${this.gte ? `-${this.gte}` : ''}`
        } else {
          transition = `${this.animation}${this.layout ? `-${this.direction}` : ''}${this.lt ? `-${this.lt}` : ''}`
        }
        this.transition = TS[transition] || TS['fade']
      }
      this.realSelected = this.index === index
    }
  }
}
</script>
