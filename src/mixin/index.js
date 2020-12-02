export default {
  methods: {
    checkSize(size) {
      if (this.smallSize != window.innerWidth <= size)
        this.smallSize = !this.smallSize
    },
    async goback(reload) {
      if (reload)
        await this.$store.dispatch('categories')
      this.$router.go(-1)
    },
    cancel(event) { if (event.key == 'Escape') this.goback() }
  }
}
