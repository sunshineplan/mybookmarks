<script lang="ts">
  import { createEventDispatcher } from "svelte";
  import { valid, confirm } from "../misc";
  import { component } from "../stores";
  import {
    category as current,
    bookmark,
    categories,
    bookmarks,
  } from "../bookmark";

  const dispatch = createEventDispatcher();

  let name = $bookmark ? $bookmark.bookmark : "";
  let url = $bookmark ? $bookmark.url : "";
  let category = $bookmark ? $bookmark.category : "";
  let validated = false;

  $: mode = window.location.pathname == "/bookmark/add" ? "Add" : "Edit";

  const chkURL = () => {
    if (url && !url.match(/^https?:/) && url.length) url = "http://" + url;
  };

  const save = async () => {
    if (valid()) {
      validated = false;
      const b = <Bookmark>{ bookmark: name, url, category };
      if (mode == "Edit") b.id = $bookmark.id;
      try {
        const res = await bookmarks.save(b);
        if (res === 0) goback();
        else if (res == 1) name = "";
        else if (res == 2) url = "";
      } catch {
        dispatch("reload");
        $current = {};
        goback();
      }
    } else validated = true;
  };

  const del = async () => {
    if (await confirm("bookmark")) {
      try {
        await bookmarks.delete($bookmark);
      } catch {
        dispatch("reload");
        $current = {};
      }
      goback();
    }
  };

  const goback = () => {
    window.history.pushState({}, "", "/");
    $component = "show";
  };
</script>

<svelte:window
  on:keydown={(e) => {
    if (e.key === "Escape") goback();
  }}
/>

<svelte:head>
  <title>{mode} Bookmark - My Bookmarks</title>
</svelte:head>

<div
  on:keydown={async (e) => {
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
      <!-- svelte-ignore a11y-autofocus -->
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
        on:blur={chkURL}
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
        {#each $categories as category (category.category)}
          <option>{category.category}</option>
        {/each}
      </datalist>
      <small class="form-text text-muted">
        Max length: 15 characters. One chinese character equal three characters.
      </small>
    </div>
    <button class="btn btn-primary" on:click={save}>{mode}</button>
    <button class="btn btn-primary" on:click={goback}>Cancel</button>
  </div>
  {#if mode == "Edit"}
    <div class="form">
      <button class="btn btn-danger delete" on:click={del}>Delete</button>
    </div>
  {/if}
</div>

<style>
  .delete {
    margin-top: 8px;
  }
</style>
