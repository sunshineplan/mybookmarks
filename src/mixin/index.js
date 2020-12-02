export default {
  methods: {
    checkSize(size) {
      if (this.smallSize != window.innerWidth <= size)
        this.smallSize = !this.smallSize
    },
    goback(reload) {
      if (reload)
        this.$store.dispatch('categories')
      this.$router.go(-1)
    },
    cancel(event) { if (event.key == 'Escape') this.goback() }
  }
}
