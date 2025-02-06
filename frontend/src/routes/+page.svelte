<script>
    import { onMount } from 'svelte';
    import { writable } from 'svelte/store';
  
    const containers = writable([]);
  
    onMount(() => {
      const socket = new WebSocket('ws://localhost:8080/ws');
      socket.addEventListener('message', (event) => {
        containers.set(JSON.parse(event.data));
      });
      return () => socket.close();
    });
  </script>
  
  <main>
    <!-- Header -->
    <h1 class="text-4xl font-extrabold text-sky-500 mb-6">🚀 DockPulse</h1>
  
    <!-- Container List -->
    <div class="bg-white shadow-lg rounded-xl p-6 w-full max-w-2xl">
      <h2 class="text-xl font-semibold text-gray-800 mb-4">🖥 Active Containers</h2>
  
      <div class="space-y-3">
        {#if $containers.length === 0}
          <p class="text-gray-500">Waiting for data...</p>
        {:else}
          {#each $containers as c}
            <div class="bg-gray-100 rounded-lg p-4 shadow flex items-center space-x-3">
              <div class="h-2 w-2 bg-green-500 rounded-full"></div>
              <div>
                <strong class="text-lime-800">{c.name}</strong>
                <p class="text-sm text-gray-600">CPU: {c.cpu} | RAM: {c.ram}</p>
                <pre class="w-fit overflow-auto whitespace-pre-wrap bg-black text-green-400 p-2 rounded-md text-xs font-mono mt-2 max-h-40">
                  {c.logs}
                </pre>
              </div>
            </div>
          {/each}
        {/if}
      </div>
    </div>
  </main>
  