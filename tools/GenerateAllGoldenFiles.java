// tools/GenerateAllGoldenFiles.java
// Generates golden hex + json for all GB/T 32960 message types.
// Build and run from the reference implementation's tools directory.
//
// Usage:
//   javac -cp <classpath> GenerateAllGoldenFiles.java
//   java -cp .:<classpath> GenerateAllGoldenFiles <output-dir>
//
// For each message type, emits:
//   <outDir>/<name>.hex  — single-line uppercase hex of encoded body bytes
//   <outDir>/<name>.json — Jackson-serialized decoded object
//
// TODO: fill in sampleXxx() factory methods with realistic field values
// matching the reference implementation's test fixtures.

import java.nio.file.*;
import java.util.LinkedHashMap;
import java.util.Map;

public class GenerateAllGoldenFiles {

    public static void main(String[] args) throws Exception {
        Path outDir = Paths.get(args.length > 0 ? args[0] : "golden/layer_c");
        Files.createDirectories(outDir);

        Map<String, Object> samples = new LinkedHashMap<>();
        // TODO: Fill in sample factories — one per message type (~50 total).
        //
        // V2016 command messages:
        //   samples.put("vehicle_login",         sampleVehicleLogin());
        //   samples.put("real_time_data_2016",   sampleRealTimeData2016());
        //   samples.put("supplement_data_2016",  sampleSupplementData2016());
        //   samples.put("vehicle_logout",        sampleVehicleLogout());
        //   samples.put("platform_login",        samplePlatformLogin());
        //   samples.put("platform_logout",       samplePlatformLogout());
        //
        // V2025 command messages:
        //   samples.put("vehicle_login_v2025",   sampleVehicleLoginV2025());
        //   samples.put("real_time_data_2025",   sampleRealTimeData2025());
        //   samples.put("supplement_data_2025",  sampleSupplementData2025());
        //   samples.put("vehicle_logout_v2025",  sampleVehicleLogoutV2025());
        //   samples.put("platform_login_v2025",  samplePlatformLoginV2025());
        //   samples.put("platform_logout_v2025", samplePlatformLogoutV2025());
        //   samples.put("vehicle_activate",      sampleVehicleActivate());
        //   samples.put("vehicle_activate_resp", sampleVehicleActivateResponse());
        //   samples.put("key_exchange",          sampleKeyExchange());
        //
        // V2016 real-time TLV sub-records:
        //   samples.put("tlv_vehicle_data_2016",    sampleVehicleData2016());
        //   samples.put("tlv_motor_data_2016",      sampleMotorData2016());
        //   samples.put("tlv_fuel_cell_2016",        sampleFuelCellData2016());
        //   samples.put("tlv_engine_data_2016",      sampleEngineData2016());
        //   samples.put("tlv_location_data_2016",    sampleLocationData2016());
        //   samples.put("tlv_extremum_2016",          sampleExtremumData2016());
        //   samples.put("tlv_alarm_data_2016",       sampleAlarmData2016());
        //   samples.put("tlv_subsystem_voltage_2016", sampleSubsystemVoltage2016());
        //   samples.put("tlv_subsystem_temp_2016",    sampleSubsystemTemp2016());
        //
        // V2025 real-time TLV sub-records:
        //   samples.put("tlv_vehicle_data_2025",     sampleVehicleData2025());
        //   samples.put("tlv_motor_data_2025",       sampleMotorData2025());
        //   samples.put("tlv_fuel_cell_engine_2025",  sampleFuelCellEngine2025());
        //   samples.put("tlv_engine_data_2025",       sampleEngineData2025());
        //   samples.put("tlv_location_data_2025",     sampleLocationData2025());
        //   samples.put("tlv_alarm_data_2025",        sampleAlarmData2025());
        //   samples.put("tlv_min_parallel_voltage_2025", sampleMinParallelVoltage2025());
        //   samples.put("tlv_battery_pack_temp_2025", sampleBatteryPackTemp2025());
        //   samples.put("tlv_fuel_cell_stack_2025",    sampleFuelCellStack2025());
        //   samples.put("tlv_super_capacitor_2025",    sampleSuperCapacitor2025());
        //   samples.put("tlv_super_cap_extremum_2025", sampleSuperCapExtremum2025());
        //   samples.put("tlv_vehicle_signature_2025",  sampleVehicleSignature2025());
        //   samples.put("tlv_custom_data_2025",        sampleCustomData2025());
        //
        // Protocol frame wrappers (header + cmd + VIN + encryption frame):
        //   samples.put("frame_vehicle_login_2016",     sampleFrameVehicleLogin2016());
        //   samples.put("frame_vehicle_login_2025",     sampleFrameVehicleLogin2025());
        //   samples.put("frame_real_time_2016",         sampleFrameRealTime2016());
        //   samples.put("frame_real_time_2025",         sampleFrameRealTime2025());
        //
        // Payload-only (prefixed payload_*.hex for the Go regression's
        // PayloadOnly sub-test):
        //   samples.put("payload_vehicle_login_2016",   samplePayloadVehicleLogin2016());
        //   samples.put("payload_vehicle_login_2025",   samplePayloadVehicleLogin2025());

        for (Map.Entry<String, Object> e : samples.entrySet()) {
            String name = e.getKey();
            Object msg = e.getValue();

            // Encode bytes via the same CodecMap the production runtime uses
            byte[] bytes = CodecMap.encode(msg);

            // Write hex (uppercase, no spaces)
            StringBuilder hex = new StringBuilder(bytes.length * 2);
            for (byte b : bytes) hex.append(String.format("%02X", b & 0xFF));
            Files.writeString(outDir.resolve(name + ".hex"), hex.toString());

            // Write json via BeanTime-aware mapper
            // MAPPER.writeValue(outDir.resolve(name + ".json").toFile(), msg);
        }
        System.out.println("Wrote " + samples.size() + " golden file pairs to " + outDir);
    }

    // ─── Sample factory stubs (to fill in) ─────────────────────────────────

    // Example of what a completed factory looks like:
    //
    // private static VehicleLogin sampleVehicleLogin() {
    //     VehicleLogin m = new VehicleLogin();
    //     m.setBeanTime(BeanTime.of(2024, 7, 31, 12, 0, 0));
    //     m.setSerialNum(1);
    //     m.setICCID("89860000000000000001");
    //     m.setCount(2);
    //     m.setLength(3);
    //     m.setCodes(Arrays.asList("001", "002"));
    //     return m;
    // }
}
