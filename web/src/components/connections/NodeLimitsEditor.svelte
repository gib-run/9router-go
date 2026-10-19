<script lang="ts">
  // Per-node capacity limits (docs/provider-limits-plan.md PR5, API in §7).
  //
  // A custom node is one upstream endpoint, so these are capacity rather than one
  // caller's allowance: how many requests per minute the endpoint tolerates, how
  // many may be in flight, how long one may queue, the node's lifetime token
  // ceiling, and the day its contract ends. 0 means unlimited everywhere, which is
  // what an operator who has not set anything sees.
  //
  // It saves through the limits endpoint alone (D8): the node's own save requires
  // name and prefix, so routing a limits edit through it would turn an empty field
  // in this form into a 400 for the whole node — and a limits save must not be able
  // to rename anything.
  import Button from '../../lib/ui/Button.svelte'
  import { api, type ProviderNodeLimits } from '../../api/client'

  interface Props {
    /** The provider id the limits belong to. */
    nodeId: string
    /** The policy the list already reported; absent when the caller has none. */
    limits?: ProviderNodeLimits | null
    /** Called after a save or a reset, with the limits the server answered. */
    onSaved?: (limits: ProviderNodeLimits) => void
  }

  let { nodeId, limits = null, onSaved }: Props = $props()

  let rpm = $state(0)
  let tpm = $state(0)
  let concurrency = $state(0)
  let tokenQuota = $state(0)
  let queueTimeoutMs = $state(0)
  let expiresOn = $state('')
  let saving = $state(false)
  let resetting = $state(false)
  let error = $state('')
  let notice = $state('')

  // Re-seed whenever the caller hands over a freshly read policy, so a save that
  // changed nothing else does not leave stale numbers in the form.
  $effect(() => {
    rpm = limits?.rpm ?? 0
    tpm = limits?.tpm ?? 0
    concurrency = limits?.concurrency ?? 0
    tokenQuota = limits?.tokenQuota ?? 0
    queueTimeoutMs = limits?.queueTimeoutMs ?? 0
    expiresOn = dateInputOf(limits?.contractExpiresAt)
  })

  function nonNegative(raw: string): number {
    const n = Number.parseInt(raw, 10)
    return Number.isFinite(n) && n > 0 ? n : 0
  }

  /** The stored RFC3339 timestamp as the date a date input shows. */
  function dateInputOf(rfc3339?: string): string {
    if (!rfc3339) return ''
    const at = new Date(rfc3339)
    return Number.isNaN(at.getTime()) ? rfc3339.slice(0, 10) : at.toISOString().slice(0, 10)
  }

  /**
   * The picked day is inclusive: an operator choosing 2027-01-31 means the node
   * works that day. Midnight is the start of the day they picked, which in a
   * positive-offset timezone is the previous evening — a paying node taken offline
   * a day early. The day's end in UTC is what gets stored.
   */
  function expiryTimestamp(date: string): string {
    return date ? `${date}T23:59:59Z` : ''
  }

  async function handleSave() {
    if (saving || resetting) return
    saving = true
    error = ''
    notice = ''
    try {
      const saved = await api.setProviderNodeLimits(nodeId, {
        rpm,
        tpm,
        concurrency,
        tokenQuota,
        queueTimeoutMs,
        contractExpiresAt: expiryTimestamp(expiresOn),
      })
      notice = 'Limits saved'
      onSaved?.(saved.limits)
    } catch (err) {
      error = err instanceof Error ? err.message : String(err)
    } finally {
      saving = false
    }
  }

  async function handleReset() {
    if (saving || resetting) return
    resetting = true
    error = ''
    notice = ''
    try {
      const out = await api.resetProviderNodeQuota(nodeId)
      notice = `Token count restarted from usage #${out.quotaResetId}`
      onSaved?.(out.limits)
    } catch (err) {
      error = err instanceof Error ? err.message : String(err)
    } finally {
      resetting = false
    }
  }
</script>

<section class="space-y-3 rounded-xl border border-border bg-surface-2 p-3">
  <div>
    <h3 class="font-headline text-xs font-bold text-text-main">Capacity limits</h3>
    <p class="text-text-subtle mt-0.5">
      Applied to every request routed through this node. <span class="font-code">0</span> means unlimited — a
      request that cannot get a slot is served by another candidate, and <span class="font-code">429</span> only
      appears once no candidate is left.
    </p>
  </div>

  <div class="grid gap-3 sm:grid-cols-2">
    <div>
      <label for="node-limits-rpm" class="block font-semibold text-text-muted mb-1">Requests / minute</label>
      <input
        id="node-limits-rpm"
        type="number"
        min="0"
        step="1"
        value={rpm}
        oninput={(e) => (rpm = nonNegative((e.target as HTMLInputElement).value))}
        class="w-full bg-surface border border-border rounded-lg px-3 py-2 font-code text-xs text-text-main focus:outline-none focus:border-primary"
      />
    </div>
    <div>
      <label for="node-limits-tpm" class="block font-semibold text-text-muted mb-1">Prompt tokens / minute</label>
      <input
        id="node-limits-tpm"
        type="number"
        min="0"
        step="1"
        value={tpm}
        oninput={(e) => (tpm = nonNegative((e.target as HTMLInputElement).value))}
        class="w-full bg-surface border border-border rounded-lg px-3 py-2 font-code text-xs text-text-main focus:outline-none focus:border-primary"
      />
    </div>
    <div>
      <label for="node-limits-conc" class="block font-semibold text-text-muted mb-1">Concurrent requests</label>
      <input
        id="node-limits-conc"
        type="number"
        min="0"
        step="1"
        value={concurrency}
        oninput={(e) => (concurrency = nonNegative((e.target as HTMLInputElement).value))}
        class="w-full bg-surface border border-border rounded-lg px-3 py-2 font-code text-xs text-text-main focus:outline-none focus:border-primary"
      />
    </div>
    <div>
      <label for="node-limits-queue" class="block font-semibold text-text-muted mb-1">Queue wait (ms)</label>
      <input
        id="node-limits-queue"
        type="number"
        min="0"
        step="1"
        value={queueTimeoutMs}
        oninput={(e) => (queueTimeoutMs = nonNegative((e.target as HTMLInputElement).value))}
        class="w-full bg-surface border border-border rounded-lg px-3 py-2 font-code text-xs text-text-main focus:outline-none focus:border-primary"
      />
      <p class="text-text-subtle mt-1">How long a request waits for a slot before the node counts as full.</p>
    </div>
    <div>
      <label for="node-limits-quota" class="block font-semibold text-text-muted mb-1">Lifetime token quota</label>
      <input
        id="node-limits-quota"
        type="number"
        min="0"
        step="1"
        value={tokenQuota}
        oninput={(e) => (tokenQuota = nonNegative((e.target as HTMLInputElement).value))}
        class="w-full bg-surface border border-border rounded-lg px-3 py-2 font-code text-xs text-text-main focus:outline-none focus:border-primary"
      />
      <p class="text-text-subtle mt-1">
        Total prompt + completion tokens for this node. Once spent, every request to it is refused with
        <span class="font-code">403</span> until the count is restarted below.
      </p>
    </div>
    <div>
      <label for="node-limits-expiry" class="block font-semibold text-text-muted mb-1">Contract ends</label>
      <input
        id="node-limits-expiry"
        type="date"
        value={expiresOn}
        oninput={(e) => (expiresOn = (e.target as HTMLInputElement).value)}
        class="w-full sm:w-48 bg-surface border border-border rounded-lg px-3 py-2 font-code text-xs text-text-main focus:outline-none focus:border-primary"
      />
      <p class="text-text-subtle mt-1">Last day the subscription is valid. Empty means no expiry.</p>
    </div>
  </div>

  {#if limits}
    <div class="rounded-lg border border-border bg-surface p-3">
      <div class="font-headline text-xs font-bold text-text-main mb-1">Usage</div>
      <div class="grid gap-1 font-code text-xs text-text-muted sm:grid-cols-3">
        <div>
          {limits.usedTokens.toLocaleString()}{#if limits.tokenQuota > 0} / {limits.tokenQuota.toLocaleString()}{/if}
          tokens
        </div>
        <div>{limits.rpmUsedWindow}{#if limits.rpm > 0} / {limits.rpm}{/if} this minute</div>
        <div>{limits.activeConcurrency}{#if limits.concurrency > 0} / {limits.concurrency}{/if} in flight</div>
      </div>
      {#if limits.expired}
        <p class="text-xs text-red-600 dark:text-red-400 mt-1">
          The contract has ended, so this node is skipped until the date is moved forward.
        </p>
      {/if}
      <div class="mt-2">
        <Button size="sm" variant="ghost" loading={resetting} onclick={handleReset}>
          Restart token count
        </Button>
      </div>
    </div>
  {/if}

  {#if notice}
    <p class="text-xs text-green-600" role="status">{notice}</p>
  {/if}
  {#if error}
    <p class="text-xs text-red-600 dark:text-red-400" role="alert">{error}</p>
  {/if}

  <Button size="sm" loading={saving} onclick={handleSave}>Save limits</Button>
</section>
