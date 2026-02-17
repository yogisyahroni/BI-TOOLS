'use client'

import * as React from 'react'
import {
  Code2,
  Copy,
  Check,
  Globe,
  Shield,
  Clock,
  Trash2,
  AlertCircle,
  Loader2,
  Plus,
  Eye,
  Lock,
  _RefreshCw,
} from 'lucide-react'
import { format } from 'date-fns'

import { cn } from '@/lib/utils'
import { Button } from '@/components/ui/button'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from '@/components/ui/dialog'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { Switch } from '@/components/ui/switch'
import { Textarea } from '@/components/ui/textarea'
import { Badge } from '@/components/ui/badge'
import { Alert, AlertDescription } from '@/components/ui/alert'
import { Separator } from '@/components/ui/separator'
import { ScrollArea } from '@/components/ui/scroll-area'
import {
  createEmbedToken,
  getEmbedTokensForResource,
  revokeEmbedToken,
} from '@/lib/api/shares'
import type {
  EmbedToken,
  ResourceType,
} from '@/types/share'

interface EmbedDialogProps {
  resourceType: ResourceType
  resourceId: string
  resourceName: string
  children?: React.ReactNode
}

export function EmbedDialog({
  resourceType,
  resourceId,
  resourceName,
  children,
}: EmbedDialogProps) {
  const [open, setOpen] = React.useState(false)
  const [tokens, setTokens] = React.useState<EmbedToken[]>([])
  const [isLoading, setIsLoading] = React.useState(false)
  const [isCreating, setIsCreating] = React.useState(false)
  const [error, setError] = React.useState<string | null>(null)
  const [copiedId, setCopiedId] = React.useState<string | null>(null)
  const [showEmbedCode, setShowEmbedCode] = React.useState<string | null>(null)

  // Form state
  const [allowedDomains, setAllowedDomains] = React.useState('')
  const [allowedIPs, setAllowedIPs] = React.useState('')
  const [setExpiration, setSetExpiration] = React.useState(false)
  const [expirationDate, setExpirationDate] = React.useState('')
  const [description, setDescription] = React.useState('')

  React.useEffect(() => {
    if (open) {
      loadTokens()
    }
        // eslint-disable-next-line react-hooks/exhaustive-deps
        // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [open])

  const loadTokens = async () => {
    setIsLoading(true)
    setError(null)
    try {
      const data = await getEmbedTokensForResource(resourceType, resourceId)
      setTokens(data)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load embed tokens')
    } finally {
      setIsLoading(false)
    }
  }

  const handleCreateToken = async (e: React.FormEvent) => {
    e.preventDefault()

    setIsCreating(true)
    setError(null)

    try {
      const request: {
        resource_type: ResourceType
        resource_id: string
        allowed_domains?: string[]
        allowed_ips?: string[]
        expires_at?: string
        description?: string
      } = {
        resource_type: resourceType,
        resource_id: resourceId,
      }

      // Parse domains (comma or newline separated)
      if (allowedDomains.trim()) {
        request.allowed_domains = allowedDomains
          .split(/[\n,]/)
          .map((d) => d.trim())
          .filter((d) => d.length > 0)
      }

      // Parse IPs (comma or newline separated)
      if (allowedIPs.trim()) {
        request.allowed_ips = allowedIPs
          .split(/[\n,]/)
          .map((ip) => ip.trim())
          .filter((ip) => ip.length > 0)
      }

      if (setExpiration && expirationDate) {
        request.expires_at = new Date(expirationDate).toISOString()
      }

      if (description.trim()) {
        request.description = description.trim()
      }

      await createEmbedToken(request)

      // Reset form
      setAllowedDomains('')
      setAllowedIPs('')
      setExpirationDate('')
      setDescription('')
      setSetExpiration(false)

      // Reload tokens
      await loadTokens()
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to create embed token')
    } finally {
      setIsCreating(false)
    }
  }

  const handleRevokeToken = async (tokenId: string) => {
    try {
      await revokeEmbedToken(tokenId)
      await loadTokens()
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to revoke token')
    }
  }

  const handleCopyToken = async (token: string, tokenId: string) => {
    try {
      await navigator.clipboard.writeText(token)
      setCopiedId(tokenId)
      setTimeout(() => setCopiedId(null), 2000)
    } catch {
      setCopiedId(tokenId)
      setTimeout(() => setCopiedId(null), 2000)
    }
  }

  const handleCopyEmbedCode = async (token: string) => {
    const embedUrl = `${window.location.origin}/embed/${token}`
    const embedCode = `<iframe
  src="${embedUrl}"
  width="100%"
  height="600"
  frameborder="0"
  allowfullscreen
></iframe>`
    try {
      await navigator.clipboard.writeText(embedCode)
      setCopiedId(`code-${token}`)
      setTimeout(() => setCopiedId(null), 2000)
    } catch {
      setCopiedId(`code-${token}`)
      setTimeout(() => setCopiedId(null), 2000)
    }
  }

  const activeTokens = tokens.filter((t) => !t.is_revoked && !isExpired(t))
  const revokedTokens = tokens.filter((t) => t.is_revoked || isExpired(t))

  function isExpired(token: EmbedToken): boolean {
    if (!token.expires_at) return false
    return new Date(token.expires_at) < new Date()
  }

  return (
    <Dialog open={open} onOpenChange={setOpen}>
      <DialogTrigger asChild>
        {children || (
          <Button variant="outline" size="sm">
            <Code2 className="mr-2 h-4 w-4" />
            Embed
          </Button>
        )}
      </DialogTrigger>
      <DialogContent className="max-w-2xl max-h-[90vh] overflow-hidden">
        <DialogHeader>
          <DialogTitle className="flex items-center gap-2">
            <Code2 className="h-5 w-5" />
            Embed {resourceType === 'dashboard' ? 'Dashboard' : 'Query'}
          </DialogTitle>
          <DialogDescription>
            Create embed tokens to share &quot;{resourceName}&quot; externally with domain and IP restrictions.
          </DialogDescription>
        </DialogHeader>

        <Tabs defaultValue="create" className="mt-4">
          <TabsList className="grid w-full grid-cols-2">
            <TabsTrigger value="create">
              <Plus className="mr-2 h-4 w-4" />
              Create Token
            </TabsTrigger>
            <TabsTrigger value="manage">
              <Shield className="mr-2 h-4 w-4" />
              Manage Tokens ({activeTokens.length})
            </TabsTrigger>
          </TabsList>

          <TabsContent value="create" className="space-y-4 mt-4">
            <form onSubmit={handleCreateToken} className="space-y-4">
              {/* Domain Restrictions */}
              <div className="space-y-2">
                <div className="flex items-center gap-2">
                  <Globe className="h-4 w-4 text-muted-foreground" />
                  <Label htmlFor="allowed-domains">Allowed Domains (Optional)</Label>
                </div>
                <Textarea
                  id="allowed-domains"
                  placeholder="example.com&#10;*.example.com&#10;app.example.com"
                  value={allowedDomains}
                  onChange={(e) => setAllowedDomains(e.target.value)}
                  rows={3}
                  className="font-mono text-sm"
                />
                <p className="text-xs text-muted-foreground">
                  Enter one domain per line or comma-separated. Use *.example.com for wildcards.
                  Leave empty to allow all domains.
                </p>
              </div>

              {/* IP Restrictions */}
              <div className="space-y-2">
                <div className="flex items-center gap-2">
                  <Lock className="h-4 w-4 text-muted-foreground" />
                  <Label htmlFor="allowed-ips">Allowed IPs (Optional)</Label>
                </div>
                <Textarea
                  id="allowed-ips"
                  placeholder="192.168.1.1&#10;10.0.0.0/24"
                  value={allowedIPs}
                  onChange={(e) => setAllowedIPs(e.target.value)}
                  rows={2}
                  className="font-mono text-sm"
                />
                <p className="text-xs text-muted-foreground">
                  Enter one IP address per line or comma-separated.
                  Leave empty to allow all IPs.
                </p>
              </div>

              {/* Expiration */}
              <div className="space-y-4 rounded-lg border p-4">
                <div className="flex items-center justify-between">
                  <div className="flex items-center gap-2">
                    <Clock className="h-4 w-4 text-muted-foreground" />
                    <Label htmlFor="set-expiration" className="cursor-pointer">
                      Set Expiration
                    </Label>
                  </div>
                  <Switch
                    id="set-expiration"
                    checked={setExpiration}
                    onCheckedChange={setSetExpiration}
                  />
                </div>

                {setExpiration && (
                  <div className="space-y-2">
                    <Label htmlFor="expiration">Expiration Date</Label>
                    <Input
                      id="expiration"
                      type="datetime-local"
                      value={expirationDate}
                      onChange={(e) => setExpirationDate(e.target.value)}
                      required={setExpiration}
                      min={new Date().toISOString().slice(0, 16)}
                    />
                  </div>
                )}
              </div>

              {/* Description */}
              <div className="space-y-2">
                <Label htmlFor="description">Description (Optional)</Label>
                <Input
                  id="description"
                  placeholder="e.g., Production website embed"
                  value={description}
                  onChange={(e) => setDescription(e.target.value)}
                />
              </div>

              {error && (
                <Alert variant="destructive">
                  <AlertCircle className="h-4 w-4" />
                  <AlertDescription>{error}</AlertDescription>
                </Alert>
              )}

              <Button
                type="submit"
                className="w-full"
                disabled={isCreating}
              >
                {isCreating ? (
                  <>
                    <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                    Creating Token...
                  </>
                ) : (
                  <>
                    <Plus className="mr-2 h-4 w-4" />
                    Create Embed Token
                  </>
                )}
              </Button>
            </form>
          </TabsContent>

          <TabsContent value="manage" className="mt-4">
            {isLoading ? (
              <div className="flex items-center justify-center py-8">
                <Loader2 className="h-6 w-6 animate-spin text-muted-foreground" />
              </div>
            ) : tokens.length === 0 ? (
              <div className="flex flex-col items-center justify-center py-8 text-center">
                <Code2 className="h-12 w-12 text-muted-foreground/50" />
                <p className="mt-4 text-sm text-muted-foreground">
                  No embed tokens yet. Create one to embed this {resourceType} externally.
                </p>
              </div>
            ) : (
              <ScrollArea className="h-[400px] pr-4">
                <div className="space-y-4">
                  {activeTokens.length > 0 && (
                    <div>
                      <h4 className="mb-2 text-sm font-medium text-muted-foreground">
                        Active Tokens
                      </h4>
                      <div className="space-y-2">
                        {activeTokens.map((token) => (
                          <TokenItem
                            key={token.id}
                            token={token}
                            onRevoke={() => handleRevokeToken(token.id)}
                            onCopy={() => handleCopyToken(token.token, token.id)}
                            onCopyCode={() => handleCopyEmbedCode(token.token)}
                            copied={copiedId === token.id}
                            codeCopied={copiedId === `code-${token.token}`}
                            showEmbedCode={showEmbedCode === token.id}
                            onToggleEmbedCode={() =>
                              setShowEmbedCode(showEmbedCode === token.id ? null : token.id)
                            }
                          />
                        ))}
                      </div>
                    </div>
                  )}

                  {revokedTokens.length > 0 && (
                    <div>
                      <Separator className="my-4" />
                      <h4 className="mb-2 text-sm font-medium text-muted-foreground">
                        Expired/Revoked
                      </h4>
                      <div className="space-y-2 opacity-60">
                        {revokedTokens.map((token) => (
                          <TokenItem
                            key={token.id}
                            token={token}
                            onRevoke={() => {}}
                            onCopy={() => {}}
                            onCopyCode={() => {}}
                            copied={false}
                            codeCopied={false}
                            showEmbedCode={false}
                            onToggleEmbedCode={() => {}}
                            disabled
                          />
                        ))}
                      </div>
                    </div>
                  )}
                </div>
              </ScrollArea>
            )}
          </TabsContent>
        </Tabs>
      </DialogContent>
    </Dialog>
  )
}

interface TokenItemProps {
  token: EmbedToken
  onRevoke: () => void
  onCopy: () => void
  onCopyCode: () => void
  copied: boolean
  codeCopied: boolean
  showEmbedCode: boolean
  onToggleEmbedCode: () => void
  disabled?: boolean
}

function TokenItem({
  token,
  onRevoke,
  onCopy,
  onCopyCode,
  copied,
  codeCopied,
  showEmbedCode,
  onToggleEmbedCode,
  disabled = false,
}: TokenItemProps) {
  const embedUrl = `${typeof window !== 'undefined' ? window.location.origin : ''}/embed/${token.token}`
  const embedCode = `<iframe
  src="${embedUrl}"
  width="100%"
  height="600"
  frameborder="0"
  allowfullscreen
></iframe>`

  const isExpired = token.expires_at
    ? new Date(token.expires_at) < new Date()
    : false

  return (
    <div
      className={cn(
        'rounded-lg border p-3 space-y-3',
        disabled && 'opacity-50'
      )}
    >
      <div className="flex items-start justify-between">
        <div className="flex items-center gap-3">
          <div className="flex h-8 w-8 items-center justify-center rounded-full bg-primary/10">
            <Code2 className="h-4 w-4" />
          </div>
          <div>
            <p className="font-medium">
              {token.description || `Embed Token ${token.token.slice(0, 8)}...`}
            </p>
            <div className="flex items-center gap-2 text-sm text-muted-foreground">
              {token.is_revoked ? (
                <Badge variant="outline">Revoked</Badge>
              ) : isExpired ? (
                <Badge variant="destructive">Expired</Badge>
              ) : (
                <Badge variant="default">Active</Badge>
              )}
              <span className="flex items-center gap-1">
                <Eye className="h-3 w-3" />
                {token.view_count} views
              </span>
              {token.expires_at && (
                <span className="flex items-center gap-1">
                  <Clock className="h-3 w-3" />
                  Expires {format(new Date(token.expires_at), 'MMM d')}
                </span>
              )}
            </div>
          </div>
        </div>

        {!disabled && (
          <div className="flex items-center gap-1">
            <Button
              variant="ghost"
              size="icon"
              onClick={onCopy}
              title="Copy token"
            >
              {copied ? (
                <Check className="h-4 w-4 text-green-500" />
              ) : (
                <Copy className="h-4 w-4" />
              )}
            </Button>
            <Button
              variant="ghost"
              size="icon"
              onClick={onRevoke}
              className="text-destructive hover:text-destructive"
              title="Revoke token"
            >
              <Trash2 className="h-4 w-4" />
            </Button>
          </div>
        )}
      </div>

      {/* Restrictions */}
      {(token.allowed_domains?.length > 0 || token.allowed_ips?.length > 0) && (
        <div className="flex flex-wrap gap-2">
          {token.allowed_domains?.length > 0 && (
            <div className="flex items-center gap-1 text-xs text-muted-foreground">
              <Globe className="h-3 w-3" />
              <span>{token.allowed_domains.length} domain(s)</span>
            </div>
          )}
          {token.allowed_ips?.length > 0 && (
            <div className="flex items-center gap-1 text-xs text-muted-foreground">
              <Lock className="h-3 w-3" />
              <span>{token.allowed_ips.length} IP(s)</span>
            </div>
          )}
        </div>
      )}

      {/* Embed Code Section */}
      {!disabled && (
        <div className="space-y-2">
          <Button
            type="button"
            variant="outline"
            size="sm"
            onClick={onToggleEmbedCode}
            className="w-full"
          >
            {showEmbedCode ? 'Hide' : 'Show'} Embed Code
          </Button>

          {showEmbedCode && (
            <div className="space-y-2">
              <div className="relative">
                <pre className="rounded-lg bg-muted p-3 text-xs overflow-x-auto">
                  <code>{embedCode}</code>
                </pre>
                <Button
                  variant="secondary"
                  size="sm"
                  onClick={onCopyCode}
                  className="absolute top-2 right-2"
                >
                  {codeCopied ? (
                    <Check className="h-3 w-3" />
                  ) : (
                    <Copy className="h-3 w-3" />
                  )}
                </Button>
              </div>
            </div>
          )}
        </div>
      )}
    </div>
  )
}
