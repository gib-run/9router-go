<script lang="ts">
  let { country }: { country: 'BR' | 'US' } = $props()
</script>

{#if country === 'BR'}
  <svg viewBox="0 0 24 16" class="h-4 w-6 shrink-0 rounded-[2px] shadow-sm" aria-label="Brazil" role="img">
    <rect width="24" height="16" rx="2" fill="#16823b" />
    <path d="M12 2.1 21 8l-9 5.9L3 8z" fill="#ffdf00" />
    <circle cx="12" cy="8" r="3.35" fill="#2456a6" />
    <path d="M9.2 7.5c1.75-.45 3.7-.2 5.45.72" fill="none" stroke="#fff" stroke-width=".55" />
  </svg>
{:else}
  <svg viewBox="0 0 24 16" class="h-4 w-6 shrink-0 rounded-[2px] shadow-sm" aria-label="United States" role="img">
    <rect width="24" height="16" rx="2" fill="#fff" />
    <path d="M0 0h24v1.23H0zm0 2.46h24v1.23H0zm0 2.46h24v1.23H0zm0 2.46h24v1.23H0zm0 2.46h24v1.23H0zm0 2.46h24v1.23H0zm0 2.46h24v1.23H0z" fill="#c62838" />
    <path d="M0 0h10.6v8H0z" fill="#244477" />
    <g fill="#fff"><circle cx="1.4" cy="1.3" r=".35"/><circle cx="4" cy="1.3" r=".35"/><circle cx="6.6" cy="1.3" r=".35"/><circle cx="9.2" cy="1.3" r=".35"/><circle cx="2.7" cy="2.7" r=".35"/><circle cx="5.3" cy="2.7" r=".35"/><circle cx="7.9" cy="2.7" r=".35"/><circle cx="1.4" cy="4.1" r=".35"/><circle cx="4" cy="4.1" r=".35"/><circle cx="6.6" cy="4.1" r=".35"/><circle cx="9.2" cy="4.1" r=".35"/><circle cx="2.7" cy="5.5" r=".35"/><circle cx="5.3" cy="5.5" r=".35"/><circle cx="7.9" cy="5.5" r=".35"/><circle cx="1.4" cy="6.9" r=".35"/><circle cx="4" cy="6.9" r=".35"/><circle cx="6.6" cy="6.9" r=".35"/><circle cx="9.2" cy="6.9" r=".35"/></g>
  </svg>
{/if}
