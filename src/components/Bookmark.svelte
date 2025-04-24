<script lang="ts">
  import { mybookmarks } from "../bookmark.svelte";
  import { confirm, valid } from "../misc.svelte";

  let name = $state(mybookmarks.bookmark.bookmark || "");
  let url = $state(mybookmarks.bookmark.url || "");
  let category = $state(mybookmarks.bookmark.category || "");
  let validated = $state(false);

  const mode = $derived(
    window.location.pathname == "/bookmark/add" ? "Add" : "Edit",
  );

  const chkURL = () => {
    if (url && !url.match(/^https?:/) && url.length) url = "http://" + url;
  };

  const save = async () => {
    if (valid()) {
      validated = false;
      const b = { bookmark: name, url, category } as Bookmark;
      if (mode == "Edit") b.id = mybookmarks.bookmark.id;
      try {
        const res = await mybookmarks.saveBookmark(b);
        if (res === 0) goback();
        else if (res == 1) name = "";
        else if (res == 2) url = "";
      } catch {
        await mybookmarks.init();
        mybookmarks.category = undefined;
        goback();
      }
    } else validated = true;
  };

  const del = async () => {
    if (await confirm("bookmark")) {
      try {
        await mybookmarks.deleteBookmark(mybookmarks.bookmark);
      } catch {
        await mybookmarks.init();
        mybookmarks.category = undefined;
      }
      goback();
    }
  };

  const goback = () => {
    window.history.pushState({}, "", "/");
    mybookmarks.component = "show";
  };
</script>

<svelte:window
  onkeydown={(e) => {
    if (e.key === "Escape") goback();
  }}
/>

<svelte:head>
  <title>{mode} Bookmark - My Bookmarks</title>
</svelte:head>

<!-- svelte-ignore a11y_no_static_element_interactions -->
<div
  onkeydown={async (e) => {
    if (e.key == "Enter") await save();
  }}
>
  <header style="padding-left: 20px">
    <h3>{mode} Bookmark</h3>
    <hr />
  </header>
  <div class="form" class:was-validated={validated}>
    <div class="mb-3">
      <label class="form-label" for="bookmark">Bookmark</label>
      <!-- svelte-ignore a11y_autofocus -->
      <input
        class="form-control"
        id="bookmark"
        bind:value={name}
        maxlength="40"
        autofocus
        required
      />
      <div class="invalid-feedback">This field is required.</div>
      <small class="form-text text-muted">Max length: 40 characters.</small>
    </div>
    <div class="mb-3">
      <label class="form-label" for="url">URL</label>
      <input
        class="form-control"
        id="url"
        type="url"
        bind:value={url}
        onblur={chkURL}
        required
      />
      <div class="invalid-feedback">Please enter a valid URL.</div>
    </div>
    <div class="mb-3">
      <label class="form-label" for="category">Category</label>
      <input
        class="form-control"
        id="category"
        list="category-list"
        bind:value={category}
        maxlength="15"
      />
      <datalist id="category-list">
        {#each mybookmarks.categories as category (category.category)}
          <option>{category.category}</option>
        {/each}
      </datalist>
      <small class="form-text text-muted">
        Max length: 15 characters. One chinese character equal three characters.
      </small>
    </div>
    <button class="btn btn-primary" onclick={save}>{mode}</button>
    <button class="btn btn-primary" onclick={goback}>Cancel</button>
  </div>
  {#if mode == "Edit"}
    <div class="form">
      <button class="btn btn-danger delete" onclick={del}>Delete</button>
    </div>
  {/if}
</div>

<style>
  .delete {
    margin-top: 8px;
  }
</style>
