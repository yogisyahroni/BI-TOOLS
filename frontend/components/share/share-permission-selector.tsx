'use client'

import * as React from 'react'
import {
  Eye,
  Edit3,
  Shield,
  Check,
} from 'lucide-react'
import { cn } from '@/lib/utils'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import type { SharePermission } from '@/types/share'

interface PermissionOption {
  value: SharePermission
  label: string
  description: string
  icon: React.ReactNode
}

const permissionOptions: PermissionOption[] = [
  {
    value: 'view',
    label: 'View only',
    description: 'Can view but not edit',
    icon: <Eye className="h-4 w-4" />,
  },
  {
    value: 'edit',
    label: 'Can edit',
    description: 'Can view and make changes',
    icon: <Edit3 className="h-4 w-4" />,
  },
  {
    value: 'admin',
    label: 'Admin',
    description: 'Full control including sharing',
    icon: <Shield className="h-4 w-4" />,
  },
]

interface SharePermissionSelectorProps {
  value: SharePermission
  onChange: (value: SharePermission) => void
  disabled?: boolean
  className?: string
}

export function SharePermissionSelector({
  value,
  onChange,
  disabled = false,
  className,
}: SharePermissionSelectorProps) {
  const selectedOption = permissionOptions.find((opt) => opt.value === value)

  return (
    <Select
      value={value}
      onValueChange={(v) => onChange(v as SharePermission)}
      disabled={disabled}
    >
      <SelectTrigger className={cn('w-[180px]', className)}>
        <SelectValue>
          {selectedOption && (
            <div className="flex items-center gap-2">
              {selectedOption.icon}
              <span>{selectedOption.label}</span>
            </div>
          )}
        </SelectValue>
      </SelectTrigger>
      <SelectContent>
        {permissionOptions.map((option) => (
          <SelectItem
            key={option.value}
            value={option.value}
            className="flex items-center gap-2"
          >
            <div className="flex items-start gap-3 py-1">
              <div className="mt-0.5 text-muted-foreground">
                {option.icon}
              </div>
              <div className="flex flex-col">
                <span className="font-medium">{option.label}</span>
                <span className="text-xs text-muted-foreground">
                  {option.description}
                </span>
              </div>
            </div>
          </SelectItem>
        ))}
      </SelectContent>
    </Select>
  )
}

interface PermissionBadgeProps {
  permission: SharePermission
  className?: string
}

export function PermissionBadge({ permission, className }: PermissionBadgeProps) {
  const config = {
    view: {
      label: 'View',
      className: 'bg-blue-100 text-blue-800 dark:bg-blue-900 dark:text-blue-200',
      icon: <Eye className="h-3 w-3" />,
    },
    edit: {
      label: 'Edit',
      className: 'bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-200',
      icon: <Edit3 className="h-3 w-3" />,
    },
    admin: {
      label: 'Admin',
      className: 'bg-purple-100 text-purple-800 dark:bg-purple-900 dark:text-purple-200',
      icon: <Shield className="h-3 w-3" />,
    },
  }

  const { label, className: badgeClassName, icon } = config[permission]

  return (
    <span
      className={cn(
        'inline-flex items-center gap-1 rounded-full px-2 py-0.5 text-xs font-medium',
        badgeClassName,
        className
      )}
    >
      {icon}
      {label}
    </span>
  )
}

interface PermissionListProps {
  value: SharePermission
  onChange: (value: SharePermission) => void
  disabled?: boolean
  className?: string
}

export function PermissionList({
  value,
  onChange,
  disabled = false,
  className,
}: PermissionListProps) {
  return (
    <div className={cn('space-y-2', className)}>
      {permissionOptions.map((option) => {
        const isSelected = value === option.value
        return (
          <button
            key={option.value}
            type="button"
            onClick={() => !disabled && onChange(option.value)}
            disabled={disabled}
            className={cn(
              'w-full flex items-start gap-3 rounded-lg border p-3 text-left transition-all',
              isSelected
                ? 'border-primary bg-primary/5'
                : 'border-border hover:border-primary/50 hover:bg-accent',
              disabled && 'opacity-50 cursor-not-allowed'
            )}
          >
            <div
              className={cn(
                'mt-0.5 flex h-5 w-5 items-center justify-center rounded-full border',
                isSelected
                  ? 'border-primary bg-primary text-primary-foreground'
                  : 'border-muted-foreground'
              )}
            >
              {isSelected && <Check className="h-3 w-3" />}
            </div>
            <div className="flex-1">
              <div className="flex items-center gap-2">
                <span className="text-muted-foreground">{option.icon}</span>
                <span className="font-medium">{option.label}</span>
              </div>
              <p className="mt-0.5 text-sm text-muted-foreground">
                {option.description}
              </p>
            </div>
          </button>
        )
      })}
    </div>
  )
}
