<script lang="ts">
  import Nav from "./components/Nav.svelte";
  import Login from "./components/Login.svelte";
  import Setting from "./components/Setting.svelte";
  import Sidebar from "./components/Sidebar.svelte";
  import Show from "./components/Show.svelte";
  import Bookmark from "./components/Bookmark.svelte";
  import {
    username,
    total,
    showSidebar,
    component,
    loading,
    categories,
    bookmarks,
    reset,
  } from "./stores";

  const getInfo = async () => {
    loading.start();
    const resp = await fetch("/info");
    loading.end();
    const info = await resp.json();
    if (Object.keys(info).length) {
      $username = info.username;
      $categories = info.categories;
      $bookmarks = info.bookmarks;
      $total = info.total;
    } else reset();
  };
  const promise = getInfo();

  const components: {
    [component: string]: typeof Setting | typeof Show | typeof Bookmark;
  } = {
    setting: Setting,
    show: Show,
    bookmark: Bookmark,
  };
</script>

<Nav bind:username={$username} on:reload={getInfo} />
{#await promise then _}
  {#if !$username}
    <Login on:info={getInfo} />
  {:else}
    <Sidebar on:reload={getInfo} />
    <div
      class="content"
      style="padding-left: 250px; opacity: {$loading ? 0.5 : 1}"
      on:mousedown={showSidebar.close}
    >
      <svelte:component this={components[$component]} on:reload={getInfo} />
    </div>
  {/if}
{/await}
<div class="loading" hidden={!$loading}>
  <div class="sk-wave sk-center">
    <div class="sk-wave-rect" />
    <div class="sk-wave-rect" />
    <div class="sk-wave-rect" />
    <div class="sk-wave-rect" />
    <div class="sk-wave-rect" />
  </div>
</div>

<style>
  .loading {
    position: fixed;
    z-index: 2;
    top: 70px;
    left: 250px;
    height: calc(100% - 70px);
    width: calc(100% - 250px);
    display: flex;
  }

  :global(:root) {
    --sk-color: #1a73e8;
  }

  :global(.content) {
    position: fixed;
    top: 0;
    padding-top: 90px;
    height: 100%;
    width: 100%;
  }

  :global(h3) {
    cursor: default;
  }

  :global(.form) {
    padding: 0 20px;
  }

  :global(.form-control) {
    width: 250px;
  }

  :global(button + button) {
    margin-left: 0.3em;
  }

  @media (max-width: 900px) {
    .loading {
      left: 0;
      width: 100%;
    }

    .content {
      padding-left: 0 !important;
    }
  }
</style>
