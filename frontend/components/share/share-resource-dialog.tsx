'use client'

import * as React from 'react'
import {
  Share2,
  Link,
  _Copy,
  Check,
  _X,
  Mail,
  User,
  Lock,
  Clock,
  Trash2,
  AlertCircle,
  Loader2,
  Shield,
  Send,
  Calendar,
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
  SharePermissionSelector,
  PermissionBadge,
} from './share-permission-selector'
import {
  createShare,
  getSharesForResource,
  revokeShare,
  _updateShare,
} from '@/lib/api/shares'
import type {
  Share,
  ResourceType,
  SharePermission,
} from '@/types/share'

interface ShareResourceDialogProps {
  resourceType: ResourceType
  resourceId: string
  resourceName: string
  children?: React.ReactNode
}

export function ShareResourceDialog({
  resourceType,
  resourceId,
  resourceName,
  children,
}: ShareResourceDialogProps) {
  const [open, setOpen] = React.useState(false)
  const [shares, setShares] = React.useState<Share[]>([])
  const [isLoading, setIsLoading] = React.useState(false)
  const [isCreating, setIsCreating] = React.useState(false)
  const [error, setError] = React.useState<string | null>(null)
  const [copiedId, setCopiedId] = React.useState<string | null>(null)

  // Form state
  const [recipientType, setRecipientType] = React.useState<'user' | 'email'>('email')
  const [recipient, setRecipient] = React.useState('')
  const [permission, setPermission] = React.useState<SharePermission>('view')
  const [requirePassword, setRequirePassword] = React.useState(false)
  const [password, setPassword] = React.useState('')
  const [setExpiration, setSetExpiration] = React.useState(false)
  const [expirationDate, setExpirationDate] = React.useState('')
  const [message, setMessage] = React.useState('')

  React.useEffect(() => {
    if (open) {
      loadShares()
    }
        // eslint-disable-next-line react-hooks/exhaustive-deps
        // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [open])

  const loadShares = async () => {
    setIsLoading(true)
    setError(null)
    try {
      const data = await getSharesForResource(resourceType, resourceId)
      setShares(data)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load shares')
    } finally {
      setIsLoading(false)
    }
  }

  const handleCreateShare = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!recipient.trim()) return

    setIsCreating(true)
    setError(null)

    try {
      const request: {
        resource_type: ResourceType
        resource_id: string
        permission: SharePermission
        shared_with?: string
        shared_email?: string
        password?: string
        expires_at?: string
        message?: string
      } = {
        resource_type: resourceType,
        resource_id: resourceId,
        permission,
      }

      if (recipientType === 'email') {
        request.shared_email = recipient.trim()
      } else {
        request.shared_with = recipient.trim()
      }

      if (requirePassword && password) {
        request.password = password
      }

      if (setExpiration && expirationDate) {
        request.expires_at = new Date(expirationDate).toISOString()
      }

      if (message.trim()) {
        request.message = message.trim()
      }

      await createShare(request)

      // Reset form
      setRecipient('')
      setPassword('')
      setExpirationDate('')
      setMessage('')
      setRequirePassword(false)
      setSetExpiration(false)

      // Reload shares
      await loadShares()
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to create share')
    } finally {
      setIsCreating(false)
    }
  }

  const handleRevokeShare = async (shareId: string) => {
    try {
      await revokeShare(shareId)
      await loadShares()
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to revoke share')
    }
  }

  const handleCopyLink = async (shareId: string) => {
    const shareUrl = `${window.location.origin}/shared/${shareId}`
    try {
      await navigator.clipboard.writeText(shareUrl)
      setCopiedId(shareId)
      setTimeout(() => setCopiedId(null), 2000)
    } catch {
      // Fallback
      setCopiedId(shareId)
      setTimeout(() => setCopiedId(null), 2000)
    }
  }

  const getShareStatusBadge = (share: Share) => {
    const statusConfig: Record<string, { label: string; variant: 'default' | 'secondary' | 'destructive' | 'outline' }> = {
      active: { label: 'Active', variant: 'default' },
      pending: { label: 'Pending', variant: 'secondary' },
      expired: { label: 'Expired', variant: 'destructive' },
      revoked: { label: 'Revoked', variant: 'outline' },
    }

    const config = statusConfig[share.status] || { label: share.status, variant: 'default' }
    return <Badge variant={config.variant}>{config.label}</Badge>
  }

  const activeShares = shares.filter((s) => s.status === 'active' || s.status === 'pending')
  const expiredShares = shares.filter((s) => s.status === 'expired' || s.status === 'revoked')

  return (
    <Dialog open={open} onOpenChange={setOpen}>
      <DialogTrigger asChild>
        {children || (
          <Button variant="outline" size="sm">
            <Share2 className="mr-2 h-4 w-4" />
            Share
          </Button>
        )}
      </DialogTrigger>
      <DialogContent className="max-w-2xl max-h-[90vh] overflow-hidden">
        <DialogHeader>
          <DialogTitle className="flex items-center gap-2">
            <Share2 className="h-5 w-5" />
            Share {resourceType === 'dashboard' ? 'Dashboard' : 'Query'}
          </DialogTitle>
          <DialogDescription>
            Share &quot;{resourceName}&quot; with others by inviting them or creating a shareable link.
          </DialogDescription>
        </DialogHeader>

        <Tabs defaultValue="invite" className="mt-4">
          <TabsList className="grid w-full grid-cols-2">
            <TabsTrigger value="invite">
              <Send className="mr-2 h-4 w-4" />
              Invite People
            </TabsTrigger>
            <TabsTrigger value="manage">
              <Shield className="mr-2 h-4 w-4" />
              Manage Access ({activeShares.length})
            </TabsTrigger>
          </TabsList>

          <TabsContent value="invite" className="space-y-4 mt-4">
            <form onSubmit={handleCreateShare} className="space-y-4">
              {/* Recipient Type */}
              <div className="flex gap-2">
                <Button
                  type="button"
                  variant={recipientType === 'email' ? 'default' : 'outline'}
                  size="sm"
                  onClick={() => setRecipientType('email')}
                  className="flex-1"
                >
                  <Mail className="mr-2 h-4 w-4" />
                  Email
                </Button>
                <Button
                  type="button"
                  variant={recipientType === 'user' ? 'default' : 'outline'}
                  size="sm"
                  onClick={() => setRecipientType('user')}
                  className="flex-1"
                >
                  <User className="mr-2 h-4 w-4" />
                  User ID
                </Button>
              </div>

              {/* Recipient Input */}
              <div className="space-y-2">
                <Label htmlFor="recipient">
                  {recipientType === 'email' ? 'Email Address' : 'User ID'}
                </Label>
                <Input
                  id="recipient"
                  type={recipientType === 'email' ? 'email' : 'text'}
                  placeholder={
                    recipientType === 'email'
                      ? 'colleague@example.com'
                      : 'Enter user ID'
                  }
                  value={recipient}
                  onChange={(e) => setRecipient(e.target.value)}
                  required
                />
              </div>

              {/* Permission */}
              <div className="space-y-2">
                <Label>Permission Level</Label>
                <SharePermissionSelector
                  value={permission}
                  onChange={setPermission}
                />
              </div>

              {/* Optional Settings */}
              <div className="space-y-4 rounded-lg border p-4">
                <div className="flex items-center justify-between">
                  <div className="flex items-center gap-2">
                    <Lock className="h-4 w-4 text-muted-foreground" />
                    <Label htmlFor="require-password" className="cursor-pointer">
                      Require Password
                    </Label>
                  </div>
                  <Switch
                    id="require-password"
                    checked={requirePassword}
                    onCheckedChange={setRequirePassword}
                  />
                </div>

                {requirePassword && (
                  <div className="space-y-2">
                    <Label htmlFor="password">Password</Label>
                    <Input
                      id="password"
                      type="password"
                      placeholder="Set a password for this share"
                      value={password}
                      onChange={(e) => setPassword(e.target.value)}
                      required={requirePassword}
                    />
                  </div>
                )}

                <Separator />

                <div className="flex items-center justify-between">
                  <div className="flex items-center gap-2">
                    <Calendar className="h-4 w-4 text-muted-foreground" />
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

              {/* Message */}
              <div className="space-y-2">
                <Label htmlFor="message">Message (Optional)</Label>
                <Textarea
                  id="message"
                  placeholder="Add a personal message..."
                  value={message}
                  onChange={(e) => setMessage(e.target.value)}
                  rows={2}
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
                disabled={!recipient.trim() || isCreating}
              >
                {isCreating ? (
                  <>
                    <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                    Sending Invite...
                  </>
                ) : (
                  <>
                    <Send className="mr-2 h-4 w-4" />
                    Send Invite
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
            ) : shares.length === 0 ? (
              <div className="flex flex-col items-center justify-center py-8 text-center">
                <Share2 className="h-12 w-12 text-muted-foreground/50" />
                <p className="mt-4 text-sm text-muted-foreground">
                  No shares yet. Invite people to collaborate on this {resourceType}.
                </p>
              </div>
            ) : (
              <ScrollArea className="h-[400px] pr-4">
                <div className="space-y-4">
                  {activeShares.length > 0 && (
                    <div>
                      <h4 className="mb-2 text-sm font-medium text-muted-foreground">
                        Active Shares
                      </h4>
                      <div className="space-y-2">
                        {activeShares.map((share) => (
                          <ShareItem
                            key={share.id}
                            share={share}
                            onRevoke={() => handleRevokeShare(share.id)}
                            onCopy={() => handleCopyLink(share.id)}
                            copied={copiedId === share.id}
                            getStatusBadge={getShareStatusBadge}
                          />
                        ))}
                      </div>
                    </div>
                  )}

                  {expiredShares.length > 0 && (
                    <div>
                      <Separator className="my-4" />
                      <h4 className="mb-2 text-sm font-medium text-muted-foreground">
                        Expired/Revoked
                      </h4>
                      <div className="space-y-2 opacity-60">
                        {expiredShares.map((share) => (
                          <ShareItem
                            key={share.id}
                            share={share}
                            onRevoke={() => {}}
                            onCopy={() => {}}
                            copied={false}
                            getStatusBadge={getShareStatusBadge}
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

interface ShareItemProps {
  share: Share
  onRevoke: () => void
  onCopy: () => void
  copied: boolean
  getStatusBadge: (share: Share) => React.ReactNode
  disabled?: boolean
}

function ShareItem({
  share,
  onRevoke,
  onCopy,
  copied,
  getStatusBadge,
  disabled = false,
}: ShareItemProps) {
  const recipient = share.shared_with_user
    ? share.shared_with_user.username || share.shared_with_user.email
    : share.shared_email || 'Unknown'

  return (
    <div
      className={cn(
        'flex items-center justify-between rounded-lg border p-3',
        disabled && 'opacity-50'
      )}
    >
      <div className="flex items-center gap-3">
        <div className="flex h-8 w-8 items-center justify-center rounded-full bg-primary/10">
          {share.shared_email ? (
            <Mail className="h-4 w-4" />
          ) : (
            <User className="h-4 w-4" />
          )}
        </div>
        <div>
          <p className="font-medium">{recipient}</p>
          <div className="flex items-center gap-2 text-sm text-muted-foreground">
            <PermissionBadge permission={share.permission} />
            {getStatusBadge(share)}
            {share.expires_at && (
              <span className="flex items-center gap-1">
                <Clock className="h-3 w-3" />
                Expires {format(new Date(share.expires_at), 'MMM d')}
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
            title="Copy share link"
          >
            {copied ? (
              <Check className="h-4 w-4 text-green-500" />
            ) : (
              <Link className="h-4 w-4" />
            )}
          </Button>
          <Button
            variant="ghost"
            size="icon"
            onClick={onRevoke}
            className="text-destructive hover:text-destructive"
            title="Revoke access"
          >
            <Trash2 className="h-4 w-4" />
          </Button>
        </div>
      )}
    </div>
  )
}
