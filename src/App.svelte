<script lang="ts">
  import type { Component } from "svelte";
  import { init } from "./bookmark";
  import Bookmark from "./components/Bookmark.svelte";
  import Login from "./components/Login.svelte";
  import Nav from "./components/Nav.svelte";
  import Setting from "./components/Setting.svelte";
  import Show from "./components/Show.svelte";
  import Sidebar from "./components/Sidebar.svelte";
  import { component, loading, showSidebar } from "./stores";

  let username = $state("");

  const load = async () => {
    loading.start();
    username = await init();
    loading.end();
  };
  const promise = load();

  const components: {
    [component: string]: Component<{
      reload: () => Promise<void>;
    }>;
  } = {
    setting: Setting,
    show: Show,
    bookmark: Bookmark,
  };
</script>

<Nav bind:username reload={load} />
{#await promise then _}
  {#if !username}
    {#if !$loading}
      <Login info={load} />
    {/if}
  {:else}
    <Sidebar />
    {@const Component = components[$component]}
    <!-- svelte-ignore a11y_no_static_element_interactions -->
    <div
      class="content"
      style="padding-left: 250px; opacity: {$loading ? 0.5 : 1}"
      onmousedown={showSidebar.close}
    >
      <Component reload={load} />
    </div>
  {/if}
{/await}
<div class={username ? "loading" : "initializing"} hidden={!$loading}>
  <div class="sk-wave sk-center">
    <div class="sk-wave-rect"></div>
    <div class="sk-wave-rect"></div>
    <div class="sk-wave-rect"></div>
    <div class="sk-wave-rect"></div>
    <div class="sk-wave-rect"></div>
  </div>
</div>

<style>
  .initializing {
    position: fixed;
    top: 70px;
    height: calc(100% - 70px);
    width: 100%;
    display: flex;
  }

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
