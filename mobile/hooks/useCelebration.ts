import { useState, useCallback } from 'react';
import { Badge } from '../types';

export type CelebrationType = 'xp' | 'levelUp' | 'badge' | 'goal';

interface XPCelebrationData {
  type: 'xp';
  amount: number;
  reason?: string;
  position?: { x: number; y: number };
}

interface LevelUpCelebrationData {
  type: 'levelUp';
  levelNumber: number;
  levelName: string;
  previousLevel?: number;
}

interface BadgeCelebrationData {
  type: 'badge';
  badge: Badge;
}

interface GoalCelebrationData {
  type: 'goal';
  goalTitle: string;
  goalDescription?: string;
  xpReward?: number;
  completedBy?: string;
  isSharedGoal?: boolean;
}

type CelebrationData =
  | XPCelebrationData
  | LevelUpCelebrationData
  | BadgeCelebrationData
  | GoalCelebrationData;

interface UseCelebrationReturn {
  currentCelebration: CelebrationData | null;
  isCelebrating: boolean;
  celebrate: (data: CelebrationData) => void;
  stopCelebration: () => void;
}

export function useCelebration(): UseCelebrationReturn {
  const [currentCelebration, setCurrentCelebration] = useState<CelebrationData | null>(null);
  const [isCelebrating, setIsCelebrating] = useState(false);

  const celebrate = useCallback((data: CelebrationData) => {
    setCurrentCelebration(data);
    setIsCelebrating(true);
  }, []);

  const stopCelebration = useCallback(() => {
    setCurrentCelebration(null);
    setIsCelebrating(false);
  }, []);

  return {
    currentCelebration,
    isCelebrating,
    celebrate,
    stopCelebration,
  };
}

export function createTodoCompletedCelebration(
  xpAmount: number = 10,
  position?: { x: number; y: number }
): XPCelebrationData {
  return {
    type: 'xp',
    amount: xpAmount,
    reason: 'Task completed!',
    position,
  };
}

export function createLevelUpCelebration(
  levelNumber: number,
  levelName: string,
  previousLevel?: number
): LevelUpCelebrationData {
  return {
    type: 'levelUp',
    levelNumber,
    levelName,
    previousLevel,
  };
}

export function createBadgeEarnedCelebration(badge: Badge): BadgeCelebrationData {
  return {
    type: 'badge',
    badge,
  };
}

export function createGoalCompletedCelebration(
  goalTitle: string,
  options: {
    goalDescription?: string;
    xpReward?: number;
    completedBy?: string;
    isSharedGoal?: boolean;
  } = {}
): GoalCelebrationData {
  return {
    type: 'goal',
    goalTitle,
    goalDescription: options.goalDescription,
    xpReward: options.xpReward ?? 50,
    completedBy: options.completedBy,
    isSharedGoal: options.isSharedGoal ?? false,
  };
}

interface XPQueueItem {
  id: string;
  amount: number;
  reason?: string;
  position?: { x: number; y: number };
}

interface UseXPQueueReturn {
  currentXP: XPQueueItem | null;
  showXP: (amount: number, reason?: string, position?: { x: number; y: number }) => void;
  onXPComplete: () => void;
}

export function useXPQueue(): UseXPQueueReturn {
  const [, setQueue] = useState<XPQueueItem[]>([]);
  const [currentXP, setCurrentXP] = useState<XPQueueItem | null>(null);
  const [isShowing, setIsShowing] = useState(false);

  const showXP = useCallback((
    amount: number,
    reason?: string,
    position?: { x: number; y: number }
  ) => {
    const newItem: XPQueueItem = {
      id: `${Date.now()}-${Math.random()}`,
      amount,
      reason,
      position,
    };

    setQueue((prev) => [...prev, newItem]);

    if (!isShowing) {
      setCurrentXP(newItem);
      setIsShowing(true);
    }
  }, [isShowing]);

  const onXPComplete = useCallback(() => {
    setQueue((prev) => {
      const remaining = prev.slice(1);
      if (remaining.length > 0) {
        setCurrentXP(remaining[0]);
        return remaining;
      } else {
        setCurrentXP(null);
        setIsShowing(false);
        return [];
      }
    });
  }, []);

  return {
    currentXP,
    showXP,
    onXPComplete,
  };
}
