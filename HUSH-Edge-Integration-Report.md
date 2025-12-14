# HUSH Edge Wallet Integration Report

**Date:** December 13, 2025

---

## Summary

We have successfully completed the initial implementation of HUSH cryptocurrency
integration into the Edge Wallet. This document details what we started with, the
design approach we followed, and all commits representing our work.

---

## Starting Point

We began with four existing repositories:

1. **HUSH Lightwalletd** (`git.hush.is/hush/lightwalletd`)
   A gRPC backend server for HUSH light wallet clients, forked from Zcash
   lightwalletd

2. **Edge Currency Accountbased** (`github.com/EdgeApp/edge-currency-accountbased`)
   Edge's plugin system for account-based cryptocurrencies including Zcash

3. **React Native Zcash** (`github.com/EdgeApp/react-native-zcash`)
   React Native wrapper for the Zcash SDK

4. **Edge React GUI** (`github.com/EdgeApp/edge-react-gui`)
   The Edge Wallet mobile application

The goal was to integrate HUSH into Edge Wallet, leveraging the fact that HUSH is
a Zcash fork using the same Sapling shielded transaction protocol and gRPC
interface.

---

## Design Approach

We followed a 4-phase implementation plan:

### Phase 1: Validate Zcash SDK Compatibility

- Confirmed HUSH lightwalletd uses the same gRPC protocol as Zcash
- Tested connectivity to HUSH lightwalletd server
- Discovered and implemented missing `GetTreeState` gRPC method required for
  checkpoint generation

### Phase 2: Create react-native-hush

- Forked `react-native-zcash` to create `react-native-hush`
- Updated package metadata and checkpoint configuration for HUSH network
- Configured checkpoint paths for HUSH mainnet
- Generated HUSH blockchain checkpoints

### Phase 3: Create Edge HUSH Plugin

- Created 5 TypeScript files implementing the HUSH currency plugin:
  - `hushInfo.ts` - Currency configuration (HUSH, port 9067, explorer URLs)
  - `hushTypes.ts` - Type definitions
  - `hushIo.ts` - Native module bridge to react-native-hush
  - `HushTools.ts` - Key derivation and address validation
  - `HushEngine.ts` - Transaction engine with sync, send, and autoshield
- Registered plugin in Edge's plugin system
- Created `rn-hush.js` entry point for React Native

### Phase 4: Integration Testing

- Built Edge Wallet Android APK with HUSH integration
- Verified gRPC connectivity with test scripts
- Successfully compiled and linked all components

### Key Technical Details

| Parameter          | Value                      |
|--------------------|----------------------------|
| Currency Code      | HUSH                       |
| Sapling Activation | Block 0 (Pure Sapling)     |
| Block Time         | 75 seconds                 |
| Default Fee        | 10,000 satoshis            |
| gRPC Port          | 9067                       |
| Explorer           | https://explorer.hush.is   |

---

## Repository Commits

All work is on the `Edge-Wallet-Intergration` branch in each repository.

### 1. DXC--edge-currency-accountbased

**URL:** https://github.com/PaulSnow/DXC--edge-currency-accountbased

| Commit     | Description                                    |
|------------|------------------------------------------------|
| `9380570f` | Add android/build to gitignore                 |
| `24c76942` | Add HUSH cryptocurrency plugin for Edge Wallet |

**Files Created:**

- `src/hush/hushInfo.ts` (64 lines)
- `src/hush/hushTypes.ts` (107 lines)
- `src/hush/hushIo.ts` (103 lines)
- `src/hush/HushTools.ts` (226 lines)
- `src/hush/HushEngine.ts` (585 lines)
- `rn-hush.js` (1 line)

---

### 2. DXC--react-native-hush

**URL:** https://github.com/PaulSnow/DXC--react-native-hush

| Commit    | Description                                    |
|-----------|------------------------------------------------|
| `64731ea` | Add HUSH gRPC test scripts                     |
| `c63ee57` | Fork react-native-zcash for HUSH cryptocurrency|

**Key Changes:**

- Package renamed to `react-native-hush`
- Checkpoint configuration updated for HUSH network
- Test scripts for validating lightwalletd connectivity

---

### 3. DXC--hush-lightwalletd

**URL:** https://github.com/PaulSnow/DXC--hush-lightwalletd

| Commit    | Description                              |
|-----------|------------------------------------------|
| `a421cfa` | Add asmap.dat for hushd peer discovery   |
| `38f9f48` | Implement GetTreeState gRPC method       |

**Key Changes:**

- Implemented `GetTreeState` in `frontend/service.go` - required for checkpoint
  generation
- Added asmap.dat for hushd autonomous system mapping

---

### 4. DXC--edge-react-gui

**URL:** https://github.com/PaulSnow/DXC--edge-react-gui

| Commit      | Description                         |
|-------------|-------------------------------------|
| `d64cb41f2` | Add HUSH cryptocurrency integration |

**Files Modified:**

- `src/util/corePlugins.ts` - Added `hush: true` to enable plugin
- `src/components/services/EdgeCoreManager.tsx` - Imported and registered
  `makeHushIo`

---

## Build Artifacts

A debug APK was successfully built:

- **Location:** `android/app/build/outputs/apk/debug/app-debug.apk`
- **Size:** 126 MB

---

## Next Steps

1. **End-to-End Testing** - Install APK on device, create HUSH wallet, test
   send/receive
2. **Checkpoint Updates** - Generate more recent checkpoints as HUSH blockchain
   syncs
3. **Public Lightwalletd** - Configure production lightwalletd server URL
4. **Pull Requests** - Submit PRs to upstream Edge repositories once testing is
   complete

---

## Technical Notes

- HUSH uses the same Sapling cryptography as Zcash, allowing SDK reuse
- The `GetTreeState` method was missing from the public HUSH lightwalletd and
  had to be implemented
- Build requires Java 17 (Java 21 has Gradle compatibility issues)
- All TypeScript compiles successfully; pre-existing errors in other plugins
  are unrelated to HUSH
