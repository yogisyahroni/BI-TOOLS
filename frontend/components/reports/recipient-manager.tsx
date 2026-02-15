'use client';

import { useState } from 'react';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';
import { Badge } from '@/components/ui/badge';
import { X, Mail, Plus, AlertCircle } from 'lucide-react';
import type { RecipientInput, RecipientType } from '@/types/scheduled-reports';

interface RecipientManagerProps {
  recipients: RecipientInput[];
  onChange: (recipients: RecipientInput[]) => void;
  disabled?: boolean;
  error?: string;
}

export function RecipientManager({
  recipients,
  onChange,
  disabled = false,
  error,
}: RecipientManagerProps) {
  const [email, setEmail] = useState('');
  const [type, setType] = useState<RecipientType>('to');
  const [emailError, setEmailError] = useState('');

  const validateEmail = (email: string): boolean => {
    const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
    return emailRegex.test(email);
  };

  const handleAdd = () => {
    if (!email.trim()) {
      setEmailError('Email is required');
      return;
    }

    if (!validateEmail(email)) {
      setEmailError('Invalid email address');
      return;
    }

    if (recipients.some((r) => r.email.toLowerCase() === email.toLowerCase())) {
      setEmailError('Email already added');
      return;
    }

    const newRecipient: RecipientInput = {
      email: email.trim(),
      type,
    };

    onChange([...recipients, newRecipient]);
    setEmail('');
    setEmailError('');
    setType('to');
  };

  const handleRemove = (index: number) => {
    const newRecipients = recipients.filter((_, i) => i !== index);
    onChange(newRecipients);
  };

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter') {
      e.preventDefault();
      handleAdd();
    }
  };

  const getRecipientColor = (type: RecipientType) => {
    switch (type) {
      case 'to':
        return 'bg-primary/10 text-primary border-primary/20';
      case 'cc':
        return 'bg-blue-50 text-blue-600 border-blue-200';
      case 'bcc':
        return 'bg-amber-50 text-amber-600 border-amber-200';
      default:
        return 'bg-muted';
    }
  };

  return (
    <div className="space-y-4">
      {/* Add Recipient Form */}
      <div className="flex gap-2">
        <div className="flex-1">
          <div className="relative">
            <Mail className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-muted-foreground" />
            <Input
              type="email"
              placeholder="Enter email address"
              value={email}
              onChange={(e) => {
                setEmail(e.target.value);
                setEmailError('');
              }}
              onKeyDown={handleKeyDown}
              disabled={disabled}
              className="pl-10"
            />
          </div>
        </div>
        <Select
          value={type}
          onValueChange={(value) => setType(value as RecipientType)}
          disabled={disabled}
        >
          <SelectTrigger className="w-24">
            <SelectValue />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="to">To</SelectItem>
            <SelectItem value="cc">CC</SelectItem>
            <SelectItem value="bcc">BCC</SelectItem>
          </SelectContent>
        </Select>
        <Button
          type="button"
          onClick={handleAdd}
          disabled={disabled}
          size="icon"
        >
          <Plus className="w-4 h-4" />
        </Button>
      </div>

      {/* Error Message */}
      {emailError && (
        <div className="flex items-center gap-2 text-sm text-destructive">
          <AlertCircle className="w-4 h-4" />
          {emailError}
        </div>
      )}

      {error && (
        <div className="flex items-center gap-2 text-sm text-destructive">
          <AlertCircle className="w-4 h-4" />
          {error}
        </div>
      )}

      {/* Recipients List */}
      {recipients.length === 0 ? (
        <div className="text-center py-8 text-muted-foreground border-2 border-dashed border-muted rounded-lg">
          <Mail className="w-8 h-8 mx-auto mb-2 opacity-50" />
          <p className="text-sm">No recipients added yet</p>
          <p className="text-xs mt-1">Add at least one recipient</p>
        </div>
      ) : (
        <div className="space-y-2">
          {/* To Recipients */}
          {recipients.filter(r => r.type === 'to').length > 0 && (
            <div>
              <Label className="text-xs text-muted-foreground mb-2 block">To</Label>
              <div className="flex flex-wrap gap-2">
                {recipients
                  .filter(r => r.type === 'to')
                  .map((recipient, index) => (
                    <Badge
                      key={`to-${index}`}
                      variant="outline"
                      className={`${getRecipientColor('to')} pl-3 pr-2 py-1`}
                    >
                      {recipient.email}
                      <button
                        type="button"
                        onClick={() => handleRemove(recipients.indexOf(recipient))}
                        disabled={disabled}
                        className="ml-2 hover:bg-primary/20 rounded p-0.5"
                      >
                        <X className="w-3 h-3" />
                      </button>
                    </Badge>
                  ))}
              </div>
            </div>
          )}

          {/* CC Recipients */}
          {recipients.filter(r => r.type === 'cc').length > 0 && (
            <div>
              <Label className="text-xs text-muted-foreground mb-2 block">CC</Label>
              <div className="flex flex-wrap gap-2">
                {recipients
                  .filter(r => r.type === 'cc')
                  .map((recipient, index) => (
                    <Badge
                      key={`cc-${index}`}
                      variant="outline"
                      className={`${getRecipientColor('cc')} pl-3 pr-2 py-1`}
                    >
                      {recipient.email}
                      <button
                        type="button"
                        onClick={() => handleRemove(recipients.indexOf(recipient))}
                        disabled={disabled}
                        className="ml-2 hover:bg-blue-100 rounded p-0.5"
                      >
                        <X className="w-3 h-3" />
                      </button>
                    </Badge>
                  ))}
              </div>
            </div>
          )}

          {/* BCC Recipients */}
          {recipients.filter(r => r.type === 'bcc').length > 0 && (
            <div>
              <Label className="text-xs text-muted-foreground mb-2 block">BCC</Label>
              <div className="flex flex-wrap gap-2">
                {recipients
                  .filter(r => r.type === 'bcc')
                  .map((recipient, index) => (
                    <Badge
                      key={`bcc-${index}`}
                      variant="outline"
                      className={`${getRecipientColor('bcc')} pl-3 pr-2 py-1`}
                    >
                      {recipient.email}
                      <button
                        type="button"
                        onClick={() => handleRemove(recipients.indexOf(recipient))}
                        disabled={disabled}
                        className="ml-2 hover:bg-amber-100 rounded p-0.5"
                      >
                        <X className="w-3 h-3" />
                      </button>
                    </Badge>
                  ))}
              </div>
            </div>
          )}
        </div>
      )}

      {/* Summary */}
      {recipients.length > 0 && (
        <div className="text-xs text-muted-foreground pt-2 border-t">
          Total: {recipients.length} recipient{recipients.length !== 1 ? 's' : ''}
          {recipients.filter(r => r.type === 'cc').length > 0 && (
            <span className="ml-2">
              ({recipients.filter(r => r.type === 'cc').length} CC)
            </span>
          )}
          {recipients.filter(r => r.type === 'bcc').length > 0 && (
            <span className="ml-2">
              ({recipients.filter(r => r.type === 'bcc').length} BCC)
            </span>
          )}
        </div>
      )}
    </div>
  );
}
