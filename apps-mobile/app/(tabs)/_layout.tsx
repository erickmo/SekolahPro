import React from 'react';
import { Tabs } from 'expo-router';
import { StyleSheet, View } from 'react-native';
import { Ionicons } from '@expo/vector-icons';
import { useSafeAreaInsets } from 'react-native-safe-area-context';

// ─── Tab Icon ─────────────────────────────────────────────────────────────────

type IoniconsName = React.ComponentProps<typeof Ionicons>['name'];

interface TabIconProps {
  name: IoniconsName;
  focused: boolean;
  color: string;
  size: number;
}

function TabIcon({ name, focused, color, size }: TabIconProps) {
  if (!focused) {
    return <Ionicons name={name} size={size} color={color} />;
  }
  return (
    <View style={tabStyles.activeIconWrapper}>
      <Ionicons name={name} size={size} color={color} />
    </View>
  );
}

const tabStyles = StyleSheet.create({
  activeIconWrapper: {
    backgroundColor: '#EEF2FF',
    borderRadius: 10,
    width: 44,
    height: 32,
    alignItems: 'center',
    justifyContent: 'center',
  },
});

// ─── Layout ───────────────────────────────────────────────────────────────────

export default function TabsLayout() {
  const insets = useSafeAreaInsets();

  return (
    <Tabs
      screenOptions={{
        headerShown: false,
        tabBarActiveTintColor: '#4F46E5',
        tabBarInactiveTintColor: '#9CA3AF',
        tabBarStyle: {
          backgroundColor: '#FFFFFF',
          borderTopWidth: 1,
          borderTopColor: '#F3F4F6',
          paddingBottom: Math.max(insets.bottom, 8),
          paddingTop: 8,
          height: 60 + Math.max(insets.bottom, 0),
          shadowColor: '#000',
          shadowOffset: { width: 0, height: -2 },
          shadowOpacity: 0.06,
          shadowRadius: 8,
          elevation: 8,
        },
        tabBarLabelStyle: {
          fontSize: 11,
          fontWeight: '600',
          marginTop: 2,
        },
        tabBarIconStyle: {
          marginBottom: -2,
        },
      }}
    >
      <Tabs.Screen
        name="index"
        options={{
          title: 'Beranda',
          tabBarIcon: ({ focused, color, size }) => (
            <TabIcon
              name={focused ? 'home' : 'home-outline'}
              focused={focused}
              color={color}
              size={size}
            />
          ),
        }}
      />
      <Tabs.Screen
        name="students"
        options={{
          title: 'Siswa',
          tabBarIcon: ({ focused, color, size }) => (
            <TabIcon
              name={focused ? 'people' : 'people-outline'}
              focused={focused}
              color={color}
              size={size}
            />
          ),
        }}
      />
      <Tabs.Screen
        name="payments"
        options={{
          title: 'Pembayaran',
          tabBarIcon: ({ focused, color, size }) => (
            <TabIcon
              name={focused ? 'card' : 'card-outline'}
              focused={focused}
              color={color}
              size={size}
            />
          ),
        }}
      />
      <Tabs.Screen
        name="health"
        options={{
          title: 'UKS',
          tabBarIcon: ({ focused, color, size }) => (
            <TabIcon
              name={focused ? 'heart' : 'heart-outline'}
              focused={focused}
              color={color}
              size={size}
            />
          ),
        }}
      />
      <Tabs.Screen
        name="library"
        options={{
          title: 'Perpustakaan',
          tabBarIcon: ({ focused, color, size }) => (
            <TabIcon
              name={focused ? 'book' : 'book-outline'}
              focused={focused}
              color={color}
              size={size}
            />
          ),
        }}
      />
      <Tabs.Screen
        name="canteen"
        options={{
          title: 'Kantin',
          tabBarIcon: ({ focused, color, size }) => (
            <TabIcon
              name={focused ? 'restaurant' : 'restaurant-outline'}
              focused={focused}
              color={color}
              size={size}
            />
          ),
        }}
      />
      <Tabs.Screen
        name="helpdesk"
        options={{
          title: 'Helpdesk',
          tabBarIcon: ({ focused, color, size }) => (
            <TabIcon
              name={focused ? 'help-circle' : 'help-circle-outline'}
              focused={focused}
              color={color}
              size={size}
            />
          ),
        }}
      />
      <Tabs.Screen
        name="alumni"
        options={{
          title: 'Alumni',
          tabBarIcon: ({ focused, color, size }) => (
            <TabIcon
              name={focused ? 'school' : 'school-outline'}
              focused={focused}
              color={color}
              size={size}
            />
          ),
        }}
      />
      <Tabs.Screen
        name="portfolio"
        options={{
          title: 'Portofolio',
          tabBarIcon: ({ focused, color, size }) => (
            <TabIcon
              name={focused ? 'folder' : 'folder-outline'}
              focused={focused}
              color={color}
              size={size}
            />
          ),
        }}
      />
      <Tabs.Screen
        name="notifications"
        options={{
          title: 'Notifikasi',
          tabBarIcon: ({ focused, color, size }) => (
            <TabIcon
              name={focused ? 'notifications' : 'notifications-outline'}
              focused={focused}
              color={color}
              size={size}
            />
          ),
        }}
      />
      <Tabs.Screen
        name="profile"
        options={{
          title: 'Profil',
          tabBarIcon: ({ focused, color, size }) => (
            <TabIcon
              name={focused ? 'person' : 'person-outline'}
              focused={focused}
              color={color}
              size={size}
            />
          ),
        }}
      />
    </Tabs>
  );
}
