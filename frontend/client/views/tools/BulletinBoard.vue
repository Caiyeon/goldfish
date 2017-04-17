<template>
  <div>
    <div class="tile is-ancestor">
    <div class="tile is-parent">
    <article class="tile is-child box">

      <div v-for="(pair, index) in bulletinPairs" class="tile is-parent is-marginless is-paddingless">

        <div class="tile is-parent is-vertical is-6">
          <article class="message" v-bind:class="pair[0].type">
            <div class="message-header">
              <p>{{ pair[0].title }}</p>
            </div>
            <div class="message-body">
              {{ pair[0].message }}
            </div>
          </article>
        </div>

        <div v-if="pair[1]" class="tile is-parent is-vertical is-6">
          <article class="message" v-bind:class="pair[1].type">
            <div class="message-header">
              <p>{{ pair[1].title }}</p>
            </div>
            <div class="message-body">
              {{ pair[1].message }}
            </div>
          </article>
        </div>

      </div>

    </article>
    </div>
    </div>
  </div>
</template>

<script>
export default {
  data () {
    return {
      csrf: '',
      bulletins: []
    }
  },

  computed: {
    bulletinPairs: function () {
      var pairs = []
      for (var i = 0; i < this.bulletins.length; i += 2) {
        try {
          pairs.push([this.bulletins[i], this.bulletins[i + 1]])
        } catch (err) {
          pairs.push([this.bulletins[i], null])
        }
      }
      return pairs
    }
  },

  mounted: function () {
    this.bulletins.push({
      'title': 'One',
      'message': 'Bacon ipsum dolor amet tri-tip tenderloin doner pig meatloaf cupim pork belly pastrami chuck salami short loin burgdoggen jowl. Ham beef capicola shank. Ham tongue jerky pork loin tail, filet mignon capicola picanha burgdoggen beef shoulder spare ribs. Jowl ground round biltong, frankfurter beef hamburger salami ball tip flank. Pork shankle jerky pastrami. Ground round tenderloin shankle bacon ham chicken turkey corned beef picanha.',
      'type': 'is-primary',
      'link': ''
    })

    this.bulletins.push({
      'title': 'Two',
      'message': 'Pork chop ham hock andouille alcatra, t-bone pig prosciutto corned beef biltong turducken. Capicola flank t-bone turkey porchetta meatloaf. Pork loin short loin spare ribs boudin. Tenderloin andouille pork chop, tail sirloin ball tip ribeye pancetta. Tri-tip ball tip porchetta pig chuck boudin. Jowl ball tip ground round shank beef. Burgdoggen venison short ribs ribeye drumstick, pork chop meatball doner pork loin porchetta tenderloin tongue.',
      'type': 'is-info',
      'link': ''
    })

    this.bulletins.push({
      'title': 'Three',
      'message': 'Bacon short loin picanha venison strip steak sausage prosciutto pancetta shankle shank frankfurter corned beef spare ribs swine. Ground round tri-tip rump frankfurter. Pig kevin chuck corned beef turkey sirloin shankle pork chop tri-tip capicola burgdoggen beef turducken short ribs drumstick. Strip steak burgdoggen meatloaf hamburger. Flank jerky venison pastrami shankle swine landjaeger porchetta pork hamburger. Ribeye capicola boudin jowl. Doner drumstick fatback capicola, tongue sirloin jowl cupim swine turducken.',
      'type': 'is-warning',
      'link': ''
    })

    this.bulletins.push({
      'title': 'Four',
      'message': 'Alcatra kielbasa jowl, pork chop boudin picanha turducken landjaeger capicola. Tenderloin t-bone spare ribs shankle, boudin beef ribs bresaola pastrami turkey tri-tip. Turkey sausage drumstick, leberkas pork belly turducken t-bone landjaeger cow. Andouille chuck ribeye boudin pork loin. Shank alcatra cow frankfurter. Chuck biltong drumstick salami frankfurter. Boudin cow ham corned beef tenderloin, drumstick ribeye doner sirloin turkey beef ribs. Spare ribs boudin fatback strip steak hamburger jerky. Pork loin boudin pork leberkas meatball tri-tip.',
      'type': 'is-danger',
      'link': ''
    })

    this.bulletins.push({
      'title': 'Five',
      'message': 'Turkey turducken ball tip, tail tongue pig kielbasa bacon. Drumstick porchetta spare ribs leberkas turducken, meatball kevin frankfurter pork burgdoggen. Cow sirloin porchetta, flank alcatra ground round corned beef kevin capicola tenderloin shoulder shank jerky. Frankfurter sausage rump porchetta kielbasa landjaeger. Sausage burgdoggen t-bone biltong hamburger, strip steak corned beef pork belly tenderloin shankle tail beef ribs filet mignon. Venison short loin brisket ham hock burgdoggen andouille, shank picanha. Flank spare ribs sausage chuck venison.',
      'type': 'is-dark',
      'link': ''
    })
  },

  methods: {
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
