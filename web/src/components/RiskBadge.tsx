import React from 'react';
import { RiskTier } from '../types';
import { ShieldAlert, AlertTriangle, AlertCircle, CheckCircle2 } from 'lucide-react';

interface Props {
  tier: RiskTier | string;
  size?: 'sm' | 'md' | 'lg';
}

export const RiskBadge: React.FC<Props> = ({ tier, size = 'md' }) => {
  const sizeClasses = {
    sm: 'text-xs px-2 py-0.5 gap-1',
    md: 'text-xs px-2.5 py-1 gap-1.5 font-medium',
    lg: 'text-sm px-3.5 py-1.5 gap-2 font-semibold',
  };

  switch (tier) {
    case 'PROHIBITED':
      return (
        <span className={`inline-flex items-center rounded-md bg-red-950/70 border border-red-500/50 text-red-400 ${sizeClasses[size]}`}>
          <ShieldAlert className="w-3.5 h-3.5 text-red-400" />
          Prohibited (Art. 5)
        </span>
      );
    case 'HIGH_RISK':
      return (
        <span className={`inline-flex items-center rounded-md bg-amber-950/70 border border-amber-500/50 text-amber-400 ${sizeClasses[size]}`}>
          <AlertTriangle className="w-3.5 h-3.5 text-amber-400" />
          High-Risk (Annex III)
        </span>
      );
    case 'LIMITED_RISK':
      return (
        <span className={`inline-flex items-center rounded-md bg-blue-950/70 border border-blue-500/50 text-blue-400 ${sizeClasses[size]}`}>
          <AlertCircle className="w-3.5 h-3.5 text-blue-400" />
          Limited Risk (Art. 50)
        </span>
      );
    default:
      return (
        <span className={`inline-flex items-center rounded-md bg-emerald-950/70 border border-emerald-500/50 text-emerald-400 ${sizeClasses[size]}`}>
          <CheckCircle2 className="w-3.5 h-3.5 text-emerald-400" />
          Minimal Risk
        </span>
      );
  }
};
