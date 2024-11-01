<script lang="ts">
  import type { Component } from "svelte";
  import { mybookmarks } from "./bookmark.svelte";
  import Bookmark from "./components/Bookmark.svelte";
  import Login from "./components/Login.svelte";
  import Nav from "./components/Nav.svelte";
  import Setting from "./components/Setting.svelte";
  import Show from "./components/Show.svelte";
  import Sidebar from "./components/Sidebar.svelte";
  import { loading } from "./misc.svelte";

  const promise = mybookmarks.init();

  const components: {
    [component: string]: Component;
  } = {
    setting: Setting,
    show: Show,
    bookmark: Bookmark,
  };

  const Content = $derived(components[mybookmarks.component]);
</script>

<Nav />
{#await promise then _}
  {#if !mybookmarks.username}
    {#if !loading.show}
      <Login />
    {/if}
  {:else}
    <Sidebar />
    <!-- svelte-ignore a11y_no_static_element_interactions -->
    <div
      class="content"
      style="padding-left: 250px; opacity: {loading.show ? 0.5 : 1}"
    >
      <Content />
    </div>
  {/if}
{/await}
<div
  class={mybookmarks.username ? "loading" : "initializing"}
  hidden={!loading.show}
>
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
