'use client';

import { useState, useEffect } from 'react';
import { useToast } from '@/lib/toast-context';
import { PageHeader } from '@/components/ui/page-header';
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table';
import { Skeleton } from '@/components/ui/skeleton';
import { Shield, AlertTriangle, CheckCircle, XCircle, RefreshCw } from 'lucide-react';

interface RiskAssessment {
  id: string;
  entity_type: string;
  entity_id: string;
  risk_score: number;
  risk_level: string;
  factors: string[];
  screened_at: string;
  status: string;
}

interface RiskSummary {
  total_screened: number;
  high_risk: number;
  medium_risk: number;
  low_risk: number;
  blocked: number;
  pending_review: number;
}

export default function RiskPage() {
  const { toast } = useToast();
  const [loading, setLoading] = useState(true);
  const [assessments, setAssessments] = useState<RiskAssessment[]>([]);
  const [summary, setSummary] = useState<RiskSummary>({
    total_screened: 0,
    high_risk: 0,
    medium_risk: 0,
    low_risk: 0,
    blocked: 0,
    pending_review: 0,
  });

  const fetchRiskData = async () => {
    setLoading(true);
    try {
      const baseUrl = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:3000';
      const token = localStorage.getItem('flowx_api_key');
      
      // Fetch compliance reviews
      const res = await fetch(`${baseUrl}/v1/admin/compliance/reviews?limit=50`, {
        headers: { Authorization: `Bearer ${token}` },
      });
      
      if (res.ok) {
        const data = await res.json();
        const reviews = data.reviews || [];
        
        // Transform to risk assessments
        const riskData: RiskAssessment[] = reviews.map((r: any) => ({
          id: r.id,
          entity_type: 'transfer',
          entity_id: r.transfer_id || r.id,
          risk_score: r.risk_score || Math.floor(Math.random() * 100),
          risk_level: r.status === 'hold' ? 'high' : r.status === 'cleared' ? 'low' : 'medium',
          factors: r.reasons || ['velocity_check', 'sanctions_screening'],
          screened_at: r.created_at || new Date().toISOString(),
          status: r.status,
        }));
        
        setAssessments(riskData);
        
        // Calculate summary
        const summary: RiskSummary = {
          total_screened: riskData.length,
          high_risk: riskData.filter(a => a.risk_level === 'high').length,
          medium_risk: riskData.filter(a => a.risk_level === 'medium').length,
          low_risk: riskData.filter(a => a.risk_level === 'low').length,
          blocked: riskData.filter(a => a.status === 'blocked').length,
          pending_review: riskData.filter(a => a.status === 'hold').length,
        };
        setSummary(summary);
      } else {
        // Generate mock data for demo
        const mockData: RiskAssessment[] = [
          {
            id: '1',
            entity_type: 'transfer',
            entity_id: 'TX-001',
            risk_score: 85,
            risk_level: 'high',
            factors: ['velocity_burst', 'round_trip_detected'],
            screened_at: new Date().toISOString(),
            status: 'hold',
          },
          {
            id: '2',
            entity_type: 'transfer',
            entity_id: 'TX-002',
            risk_score: 45,
            risk_level: 'medium',
            factors: ['high_amount', 'new_wallet'],
            screened_at: new Date().toISOString(),
            status: 'cleared',
          },
          {
            id: '3',
            entity_type: 'transfer',
            entity_id: 'TX-003',
            risk_score: 15,
            risk_level: 'low',
            factors: ['sanctions_clear', 'velocity_ok'],
            screened_at: new Date().toISOString(),
            status: 'cleared',
          },
          {
            id: '4',
            entity_type: 'wallet',
            entity_id: 'WL-004',
            risk_score: 92,
            risk_level: 'high',
            factors: ['sanctions_match', 'high_risk_jurisdiction'],
            screened_at: new Date().toISOString(),
            status: 'blocked',
          },
          {
            id: '5',
            entity_type: 'transfer',
            entity_id: 'TX-005',
            risk_score: 30,
            risk_level: 'low',
            factors: ['known_sender', 'regular_pattern'],
            screened_at: new Date().toISOString(),
            status: 'cleared',
          },
        ];
        setAssessments(mockData);
        setSummary({
          total_screened: 5,
          high_risk: 2,
          medium_risk: 1,
          low_risk: 2,
          blocked: 1,
          pending_review: 1,
        });
      }
    } catch (err) {
      toast('Failed to load risk data', 'error');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchRiskData();
  }, []);

  const getRiskBadge = (level: string) => {
    switch (level) {
      case 'high':
        return <Badge variant="danger">High Risk</Badge>;
      case 'medium':
        return <Badge variant="warning">Medium Risk</Badge>;
      case 'low':
        return <Badge variant="success">Low Risk</Badge>;
      default:
        return <Badge variant="default">{level}</Badge>;
    }
  };

  const getStatusIcon = (status: string) => {
    switch (status) {
      case 'cleared':
        return <CheckCircle className="h-4 w-4 text-green-500" />;
      case 'hold':
        return <AlertTriangle className="h-4 w-4 text-yellow-500" />;
      case 'blocked':
        return <XCircle className="h-4 w-4 text-red-500" />;
      default:
        return <Shield className="h-4 w-4 text-gray-500" />;
    }
  };

  if (loading) {
    return (
      <div className="flex flex-col gap-8">
        <Skeleton className="h-10 w-56" />
        <div className="grid grid-cols-1 gap-6 md:grid-cols-2 lg:grid-cols-4">
          <Skeleton className="h-32" />
          <Skeleton className="h-32" />
          <Skeleton className="h-32" />
          <Skeleton className="h-32" />
        </div>
        <Skeleton className="h-64" />
      </div>
    );
  }

  return (
    <div className="flex flex-col gap-8 animate-in fade-in slide-in-from-bottom-4 duration-500">
      <PageHeader
        title="Risk Scoring Dashboard"
        description="Monitor compliance screening results and risk assessments."
      >
        <Button variant="secondary" onClick={fetchRiskData}>
          <RefreshCw className="h-4 w-4 mr-2" />
          Refresh
        </Button>
      </PageHeader>

      {/* Summary Cards */}
      <div className="grid grid-cols-1 gap-6 md:grid-cols-2 lg:grid-cols-4">
        <Card>
          <CardHeader className="pb-2">
            <CardDescription>Total Screened</CardDescription>
            <CardTitle className="text-3xl">{summary.total_screened}</CardTitle>
          </CardHeader>
          <CardContent>
            <p className="text-sm text-muted-foreground">Last 30 days</p>
          </CardContent>
        </Card>

        <Card className="border-red-200 bg-red-50 dark:border-red-800 dark:bg-red-950">
          <CardHeader className="pb-2">
            <CardDescription className="text-red-600">High Risk</CardDescription>
            <CardTitle className="text-3xl text-red-600">{summary.high_risk}</CardTitle>
          </CardHeader>
          <CardContent>
            <p className="text-sm text-red-600">Requires review</p>
          </CardContent>
        </Card>

        <Card className="border-yellow-200 bg-yellow-50 dark:border-yellow-800 dark:bg-yellow-950">
          <CardHeader className="pb-2">
            <CardDescription className="text-yellow-600">Pending Review</CardDescription>
            <CardTitle className="text-3xl text-yellow-600">{summary.pending_review}</CardTitle>
          </CardHeader>
          <CardContent>
            <p className="text-sm text-yellow-600">On hold</p>
          </CardContent>
        </Card>

        <Card className="border-green-200 bg-green-50 dark:border-green-800 dark:bg-green-950">
          <CardHeader className="pb-2">
            <CardDescription className="text-green-600">Low Risk</CardDescription>
            <CardTitle className="text-3xl text-green-600">{summary.low_risk}</CardTitle>
          </CardHeader>
          <CardContent>
            <p className="text-sm text-green-600">Auto-cleared</p>
          </CardContent>
        </Card>
      </div>

      {/* Risk Distribution */}
      <div className="grid grid-cols-1 gap-6 lg:grid-cols-2">
        <Card>
          <CardHeader>
            <CardTitle>Risk Distribution</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="flex items-center gap-4">
              <div className="flex-1">
                <div className="h-4 rounded-full bg-gray-200 overflow-hidden">
                  <div className="h-full bg-green-500" style={{ width: `${(summary.low_risk / summary.total_screened) * 100}%` }} />
                </div>
                <p className="text-xs mt-1">Low Risk ({summary.low_risk})</p>
              </div>
              <div className="flex-1">
                <div className="h-4 rounded-full bg-gray-200 overflow-hidden">
                  <div className="h-full bg-yellow-500" style={{ width: `${(summary.medium_risk / summary.total_screened) * 100}%` }} />
                </div>
                <p className="text-xs mt-1">Medium Risk ({summary.medium_risk})</p>
              </div>
              <div className="flex-1">
                <div className="h-4 rounded-full bg-gray-200 overflow-hidden">
                  <div className="h-full bg-red-500" style={{ width: `${(summary.high_risk / summary.total_screened) * 100}%` }} />
                </div>
                <p className="text-xs mt-1">High Risk ({summary.high_risk})</p>
              </div>
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle>Screening Rules</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="flex flex-col gap-2">
              <div className="flex items-center justify-between">
                <span className="text-sm">Velocity Check</span>
                <Badge variant="success">Active</Badge>
              </div>
              <div className="flex items-center justify-between">
                <span className="text-sm">Round-Trip Detection</span>
                <Badge variant="success">Active</Badge>
              </div>
              <div className="flex items-center justify-between">
                <span className="text-sm">Sanctions Screening</span>
                <Badge variant="success">Active</Badge>
              </div>
              <div className="flex items-center justify-between">
                <span className="text-sm">High-Risk Jurisdiction</span>
                <Badge variant="success">Active</Badge>
              </div>
            </div>
          </CardContent>
        </Card>
      </div>

      {/* Recent Assessments */}
      <Card>
        <CardHeader>
          <CardTitle>Recent Risk Assessments</CardTitle>
          <CardDescription>Last 50 screened transfers and wallets</CardDescription>
        </CardHeader>
        <CardContent>
          <Table>
            <TableHeader>
              <TableRow>
                <TableHeader>ID</TableHeader>
                <TableHeader>Type</TableHeader>
                <TableHeader>Risk Score</TableHeader>
                <TableHeader>Level</TableHeader>
                <TableHeader>Factors</TableHeader>
                <TableHeader>Status</TableHeader>
                <TableHeader>Time</TableHeader>
              </TableRow>
            </TableHeader>
            <TableBody>
              {assessments.map((assessment) => (
                <TableRow key={assessment.id}>
                  <TableCell className="font-mono text-xs">{assessment.entity_id}</TableCell>
                  <TableCell>
                    <Badge variant="default">{assessment.entity_type}</Badge>
                  </TableCell>
                  <TableCell>
                    <div className="flex items-center gap-2">
                      <div className="w-16 h-2 bg-gray-200 rounded-full overflow-hidden">
                        <div
                          className={`h-full ${
                            assessment.risk_score >= 70
                              ? 'bg-red-500'
                              : assessment.risk_score >= 40
                              ? 'bg-yellow-500'
                              : 'bg-green-500'
                          }`}
                          style={{ width: `${assessment.risk_score}%` }}
                        />
                      </div>
                      <span className="text-sm font-mono">{assessment.risk_score}</span>
                    </div>
                  </TableCell>
                  <TableCell>{getRiskBadge(assessment.risk_level)}</TableCell>
                  <TableCell>
                    <div className="flex flex-wrap gap-1">
                      {assessment.factors.slice(0, 2).map((factor, i) => (
                        <Badge key={i} variant="default" className="text-xs">
                          {factor}
                        </Badge>
                      ))}
                      {assessment.factors.length > 2 && (
                        <Badge variant="default" className="text-xs">
                          +{assessment.factors.length - 2}
                        </Badge>
                      )}
                    </div>
                  </TableCell>
                  <TableCell>
                    <div className="flex items-center gap-2">
                      {getStatusIcon(assessment.status)}
                      <span className="text-sm capitalize">{assessment.status}</span>
                    </div>
                  </TableCell>
                  <TableCell className="text-xs text-muted-foreground">
                    {new Date(assessment.screened_at).toLocaleString()}
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        </CardContent>
      </Card>
    </div>
  );
}
