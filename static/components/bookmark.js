const bookmark = {
  props: {
    bookmark: Object,
    categories: Array
  },
  data() {
    return {
      name: this.bookmark.Name,
      url: this.bookmark.URL,
      category: this.bookmark.Category,
      validated: false
    }
  },
  template: `
  <div>
    <header style='padding-left: 20px'>
      <a class='h3 title'>{{ mode }} Bookmark</a>
      <hr>
    </header>
    <div class='form' :class="{ 'was-validated': validated }">
      <div class='form-group'>
        <label for='bookmark'>Bookmark</label>
        <input class='form-control' v-model='name' id='bookmark' maxlength=40 required>
        <div class='invalid-feedback'>This field is required.</div>
        <small class='form-text text-muted'>Max length: 40 characters.</small>
      </div>
      <div class='form-group'>
        <label for='url'>URL</label>
        <input class='form-control' type='url' v-model='url' id='url' @blur='chkURL' required>
        <div class='invalid-feedback'>Please enter a valid URL.</div>
      </div>
      <div class='form-group'>
        <label for='category'>Category</label>
        <input class='form-control' list='category-list' v-model='category' id='category' maxlength=15>
        <datalist id='category-list'>
          <option v-for='c in categories'>{{ c.Name }}</option>
        </datalist>
        <small class='form-text text-muted'>Max length: 15 characters. One chinese character equal three characters.</small>
      </div>
      <button class='btn btn-primary' @click='save'>{{ mode }}</button>
      <button class='btn btn-primary' @click='goback'>Cancel</button>
    </div>
    <div class='form' v-if='bookmark.ID != undefined'>
      <button class='btn btn-danger delete' @click='del'>Delete</button>
    </div>
  </div>`,
  mounted() { document.title = this.mode + ' Bookmark - My Bookmarks' },
  methods: {
    chkURL: function () {
      if (!this.url.match(/^https?:/) && this.url.length)
        this.url = 'http://' + this.url
    },
    save: function () {
      if (valid()) {
        this.validated = false
        var r
        if (this.bookmark.ID == undefined)
          r = post('/bookmark/add', {
            bookmark: this.name,
            url: this.url
          })
        else
          r = post('/bookmark/edit/' + this.bookmark.ID, {
            bookmark: this.name,
            url: this.url
          })
        r.then(resp => {
          if (!resp.ok) resp.text().then(err =>
            BootstrapButtons.fire('Error', err, 'error'))
          else resp.json().then(json => {
            if (json.status == 1) this.$parent.content = 'showBookmark'
            else BootstrapButtons.fire('Error', json.message, 'error')
              .then(() => {
                if (json.error == 1) this.name = ''
                else if (json.error == 2) this.url = ''
              })
          })
        }).catch(e => BootstrapButtons.fire('Error', e, 'error'))
      }
      else this.validated = true
    },
    del: function () {
      confirm('category').then(confirm => {
        if (confirm) post('/bookmark/delete/' + this.bookmark.ID)
          .then(resp => {
            if (!resp.ok) resp.text().then(err =>
              BootstrapButtons.fire('Error', err, 'error'))
            else this.$parent.content = 'showBookmark'
          })
          .catch(e => BootstrapButtons.fire('Error', e, 'error'))
      })
    },
    goback: function () { this.$parent.content = 'showBookmark' }
  },
  computed: {
    mode: function () {
      if (this.bookmark.ID == undefined) return 'Add'
      return 'Edit'
    }
  }
}
