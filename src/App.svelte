<script lang="ts">
  import Nav from "./components/Nav.svelte";
  import Login from "./components/Login.svelte";
  import Setting from "./components/Setting.svelte";
  import Sidebar from "./components/Sidebar.svelte";
  import Show from "./components/Show.svelte";
  import Bookmark from "./components/Bookmark.svelte";
  import {
    username,
    showSidebar,
    component,
    loading,
    categories,
  } from "./stores";

  const getInfo = async () => {
    const resp = await fetch("/info");
    const info = await resp.json();
    if (Object.keys(info).length) {
      $username = info.username;
      $categories = info.categories;
    }
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

<Nav bind:username={$username} />
{#await promise then _}
  {#if !$username}
    <Login on:info={getInfo} />
  {:else}
    <Sidebar on:reload={getInfo} />
    <div
      class="content"
      style="padding-left: 250px; opacity: {$loading ? 0.5 : 1}"
      on:mousedown={() => ($showSidebar = false)}
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

  @media (max-width: 900px) {
    .loading {
      left: 0;
      width: 100%;
    }
  }
</style>
