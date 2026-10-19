<script lang="ts">
  // Port of upstream AddCustomModelModal.js (dashboard/providers/[id]).
  import Button from '../../lib/ui/Button.svelte'
  import Modal from '../../lib/ui/Modal.svelte'
  import Toggle from '../../lib/ui/Toggle.svelte'
  import { api, type CustomModelBatchEntry } from '../../api/client'
  import { parseModelIds, summarizeBulkModels } from '../../lib/bulk-models'

  export interface ModelCaps {
    vision: boolean
    reasoning: boolean
  }

  /** The operator's own declaration of what the endpoint takes. 0 = unknown. */
  export interface ModelLimits {
    contextWindow: number
    maxOutput: number
  }

  interface Props {
    isOpen: boolean
    /** Storage alias of the provider (e.g. "oc") — used to strip the alias prefix and to test. */
    providerAlias: string
    onSave: (modelId: string, caps: ModelCaps, limits: ModelLimits) => Promise<void> | void
    /** Called after a bulk save, so the parent reloads the list the paste changed. */
    onBulkSaved?: () => void
    onClose: () => void
  }

  let { isOpen, providerAlias, onSave, onBulkSaved, onClose }: Props = $props()

  // Port of CAPACITY_META (upstream shared/constants/models.js).
  const CAPACITY_META: Record<string, { label: string; desc: string }> = {
    vision: { label: 'Vision', desc: 'Supports image input' },
    reasoning: { label: 'Reasoning', desc: 'Supports reasoning / thinking' },
  }

  const defaultCaps = (): ModelCaps => ({ vision: false, reasoning: false })
  const defaultLimits = (): ModelLimits => ({ contextWindow: 0, maxOutput: 0 })

  let modelId = $state('')
  let caps = $state<ModelCaps>(defaultCaps())
  let limits = $state<ModelLimits>(defaultLimits())
  let testStatus = $state<'testing' | 'ok' | 'error' | null>(null)
  let testError = $state('')
  let saving = $state(false)

  // Bulk mode: one paste, one request (plan D9). The split/dedupe rules live in the
  // parser with their own test rather than in the markup, because a paste arrives
  // in whatever shape the provider page used.
  let mode = $state<'single' | 'bulk'>('single')
  let bulkText = $state('')
  let bulkSaving = $state(false)
  let bulkError = $state('')
  let bulkResult = $state<{ added: string[]; skipped: Array<{ id: string; reason: string; code: string }> } | null>(
    null,
  )
  let bulkPlan = $derived(parseModelIds(bulkText, providerAlias))

  function parseLimit(raw: string): number {
    const n = Number.parseInt(raw, 10)
    return Number.isFinite(n) && n > 0 ? n : 0
  }

  // max_output_tokens is drawn from the same window, so declaring both with
  // the output larger than the window is a contradiction rather than a value
  // the gateway can publish.
  let limitsError = $derived(
    limits.contextWindow > 0 && limits.maxOutput > limits.contextWindow
      ? 'Max output cannot exceed the context window.'
      : '',
  )

  // Reset state when modal opens (upstream parity).
  $effect(() => {
    if (isOpen) {
      modelId = ''
      caps = defaultCaps()
      limits = defaultLimits()
      testStatus = null
      testError = ''
      mode = 'single'
      bulkText = ''
      bulkResult = null
      bulkError = ''
    }
  })

  /** Strip provider's own alias prefix (e.g. "cc/model" -> "model" for cc provider). */
  function stripAlias(id: string): string {
    const prefix = `${providerAlias}/`
    return id.startsWith(prefix) ? id.slice(prefix.length) : id
  }

  let cleanId = $derived(stripAlias(modelId.trim()))

  async function handleTest() {
    if (!cleanId || testStatus === 'testing') return
    testStatus = 'testing'
    testError = ''
    try {
      const data = await api.testModel(`${providerAlias}/${cleanId}`)
      testStatus = data.ok ? 'ok' : 'error'
      testError = data.error || ''
    } catch (err) {
      testStatus = 'error'
      testError = err instanceof Error ? err.message : String(err)
    }
  }

  async function handleSave() {
    if (!cleanId || saving || limitsError) return
    saving = true
    try {
      await onSave(cleanId, { ...caps }, { ...limits })
    } finally {
      saving = false
    }
  }

  async function handleBulkSave() {
    if (bulkSaving || bulkPlan.ids.length === 0) return
    bulkSaving = true
    bulkError = ''
    try {
      // The response names every id it stored and every id it refused, so the counts
      // below are what landed instead of what was asked for. One entry per id, each
      // carrying the same capabilities: the form applies them to the whole paste
      // (D10), and the endpoint stays general for a caller that does not.
      bulkResult = await api.saveCustomModelsBulk({
        providerAlias,
        models: bulkPlan.ids.map((id): CustomModelBatchEntry => ({ id, ...caps, ...limits })),
      })
      onBulkSaved?.()
    } catch (err) {
      bulkResult = null
      bulkError = err instanceof Error ? err.message : String(err)
    } finally {
      bulkSaving = false
    }
  }

  function handleKeyDown(e: KeyboardEvent) {
    if (e.key === 'Enter') {
      e.preventDefault()
      handleTest()
    }
  }
</script>

<Modal {isOpen} {onClose} title="Add Custom Model">
  <div class="flex flex-col gap-4">
    <div class="flex gap-2">
      <Button
        size="sm"
        variant={mode === 'single' ? 'primary' : 'ghost'}
        onclick={() => {
          mode = 'single'
          bulkResult = null
        }}
      >
        Single
      </Button>
      <Button
        size="sm"
        variant={mode === 'bulk' ? 'primary' : 'ghost'}
        onclick={() => {
          mode = 'bulk'
          bulkResult = null
        }}
      >
        Bulk Add
      </Button>
    </div>
    {#if mode === 'single'}
    <div>
      <label class="text-sm font-medium mb-1.5 block" for="add-custom-model-id">Model ID</label>
      <div class="flex gap-2">
        <input
          id="add-custom-model-id"
          type="text"
          value={modelId}
          oninput={(e) => {
            modelId = (e.target as HTMLInputElement).value
            testStatus = null
            testError = ''
          }}
          onkeydown={handleKeyDown}
          placeholder="e.g. claude-opus-4-5"
          class="flex-1 px-3 py-2 text-sm border border-border rounded-lg bg-background focus:outline-none focus:border-primary"
        />
        <Button
          variant="secondary"
          loading={testStatus === 'testing'}
          onclick={handleTest}
          disabled={!modelId.trim() || testStatus === 'testing'}
        >
          {#if testStatus !== 'testing'}
            <span class="material-symbols-outlined text-[18px]">science</span>
          {/if}
          {testStatus === 'testing' ? 'Testing...' : 'Test'}
        </Button>
      </div>
      <p class="text-xs text-text-muted mt-1">
        Sent to provider as:
        <code class="font-mono bg-sidebar px-1 rounded">{cleanId || 'model-id'}</code>
      </p>
    </div>
    {/if}

    {#if mode === 'bulk'}
      <div>
        <label class="text-sm font-medium mb-1.5 block" for="add-custom-model-bulk">Model IDs</label>
        <textarea
          id="add-custom-model-bulk"
          rows="6"
          value={bulkText}
          oninput={(e) => {
            bulkText = (e.target as HTMLTextAreaElement).value
            bulkResult = null
          }}
          placeholder={'glm-5.3, glm-4.7\nclaude-sonnet-4.5 gpt-5.2'}
          class="w-full px-3 py-2 text-sm font-mono border border-border rounded-lg bg-background focus:outline-none focus:border-primary"
        ></textarea>
        <p class="text-xs text-text-muted mt-1">
          Separated by spaces, commas or newlines. A
          <code class="font-mono bg-sidebar px-1 rounded">{providerAlias}/</code> prefix is stripped, and the
          capabilities and token limits below apply to every id in the paste.
        </p>
        {#if bulkPlan.duplicates.length > 0 || bulkPlan.invalid.length > 0}
          <p class="text-xs text-text-muted mt-1">
            {bulkPlan.ids.length} to add
            {#if bulkPlan.duplicates.length > 0}· {bulkPlan.duplicates.length} repeated in the paste{/if}
            {#if bulkPlan.invalid.length > 0}· {bulkPlan.invalid.length} invalid{/if}
          </p>
        {/if}
      </div>

      {#if bulkResult}
        <div class="text-sm font-medium {bulkResult.skipped.length > 0 ? 'text-yellow-400' : 'text-green-400'}">
          ✓ {summarizeBulkModels(bulkPlan, bulkResult.added.length)}
        </div>
        {#if bulkResult.skipped.length > 0}
          <ul class="text-xs text-text-muted space-y-0.5">
            {#each bulkResult.skipped as skip (skip.id)}
              <li>
                <code class="font-mono">{skip.id}</code>
                — {skip.reason}
              </li>
            {/each}
          </ul>
        {/if}
      {/if}
      {#if bulkError}
        <p class="text-xs text-red-600 dark:text-red-400" role="alert">{bulkError}</p>
      {/if}
    {/if}

    <div>
      <label class="text-sm font-medium mb-1.5 block">Capabilities</label>
      <div class="flex flex-wrap gap-4">
        {#each Object.entries(CAPACITY_META) as [key, meta]}
          <div class="flex items-center gap-2">
            <Toggle
              checked={caps[key as keyof ModelCaps]}
              onchange={(v: boolean) => {
                caps = { ...caps, [key]: v }
              }}
              label={meta.label}
              size="sm"
            />
            <div class="leading-tight">
              <div class="text-sm">{meta.label}</div>
              <div class="text-xs text-text-muted">{meta.desc}</div>
            </div>
          </div>
        {/each}
      </div>
    </div>

    <div>
      <label class="text-sm font-medium mb-1.5 block">Token limits (optional)</label>
      <div class="grid gap-3 sm:grid-cols-2">
        <div>
          <label class="text-xs text-text-muted mb-1 block" for="add-custom-model-ctx">Context window</label>
          <input
            id="add-custom-model-ctx"
            type="text"
            inputmode="numeric"
            placeholder="e.g. 131072"
            value={limits.contextWindow || ''}
            oninput={(e) => {
              limits = { ...limits, contextWindow: parseLimit((e.target as HTMLInputElement).value) }
            }}
            class="w-full px-3 py-2 text-sm border border-border rounded-lg bg-background focus:outline-none focus:border-primary"
          />
        </div>
        <div>
          <label class="text-xs text-text-muted mb-1 block" for="add-custom-model-out">Max output</label>
          <input
            id="add-custom-model-out"
            type="text"
            inputmode="numeric"
            placeholder="e.g. 8192"
            value={limits.maxOutput || ''}
            oninput={(e) => {
              limits = { ...limits, maxOutput: parseLimit((e.target as HTMLInputElement).value) }
            }}
            aria-describedby={limitsError ? 'add-custom-model-limits-error' : undefined}
            aria-invalid={limitsError ? 'true' : undefined}
            class="w-full px-3 py-2 text-sm border border-border rounded-lg bg-background focus:outline-none focus:border-primary"
          />
        </div>
      </div>
      <p class="text-xs text-text-muted mt-1">
        Published to <code class="font-mono bg-sidebar px-1 rounded">/v1/models</code> instead of a guessed
        value. Leave blank to keep the default.
      </p>
      {#if limitsError}
        <p id="add-custom-model-limits-error" class="text-xs text-red-600 dark:text-red-400 mt-1" role="alert">
          {limitsError}
        </p>
      {/if}
    </div>

    {#if testStatus === 'ok'}
      <div class="flex items-center gap-2 text-sm text-green-600">
        <span class="material-symbols-outlined text-base">check_circle</span>
        Model is reachable
      </div>
    {/if}
    {#if testStatus === 'error'}
      <div class="flex items-start gap-2 text-sm text-red-500">
        <span class="material-symbols-outlined text-base shrink-0">cancel</span>
        <span>{testError || 'Model not reachable'}</span>
      </div>
    {/if}

    <div class="flex gap-2 pt-1">
      <Button onclick={onClose} variant="ghost" fullWidth size="sm">Cancel</Button>
      {#if mode === 'single'}
        <Button onclick={handleSave} fullWidth size="sm" disabled={!modelId.trim() || saving || !!limitsError}>
          {saving ? 'Adding...' : 'Add Model'}
        </Button>
      {:else}
        <Button
          onclick={handleBulkSave}
          fullWidth
          size="sm"
          disabled={bulkPlan.ids.length === 0 || bulkSaving}
        >
          {bulkSaving ? 'Adding...' : `Add ${bulkPlan.ids.length} Model${bulkPlan.ids.length === 1 ? '' : 's'}`}
        </Button>
      {/if}
    </div>
  </div>
</Modal>
