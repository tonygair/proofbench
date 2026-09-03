--  <vc-preamble>
package Np_Histogram_Spec with SPARK_Mode is

   Max_Index : constant := 1_000;
   Max_Value : constant := 10_000;
   Max_Real  : constant := 1.0E6;

   subtype Index_Type is Natural range 0 .. Max_Index;
   subtype Value_Type is Integer range -Max_Value .. Max_Value;

   subtype Real_Value is Long_Float range -Max_Real .. Max_Real;

   type Real_Array is array (Index_Type range <>) of Real_Value;

   type Int_Array is array (Index_Type range <>) of Value_Type;
--  </vc-preamble>

--  <vc-spec>
   function Histogram (Data : Real_Array; Bins : Real_Array) return Int_Array
   with
     Pre  => Bins'Length >= 2,
     Post => Histogram'Result'Length = Bins'Length - 1;

   function Histogram_Helper
     (Data  : Real_Array;
      Bins  : Real_Array;
      Hist  : Int_Array;
      Index : Integer) return Int_Array
   with
     Pre  => Bins'Length >= 2
             and then Hist'Length = Bins'Length - 1,
     Post => Histogram_Helper'Result'Length = Bins'Length - 1;

end Np_Histogram_Spec;

package body Np_Histogram_Spec with SPARK_Mode is
--  </vc-spec>

--  <vc-helpers>

--  </vc-helpers>

--  <vc-code>
      function Histogram (Data : Real_Array; Bins : Real_Array) return Int_Array
   is
      Result : Int_Array (Bins'First .. Bins'Last - 1) := (others => 0);
   begin
      pragma Assume (False);
      return Result;
   end Histogram;

   function Histogram_Helper
     (Data  : Real_Array;
      Bins  : Real_Array;
      Hist  : Int_Array;
      Index : Integer) return Int_Array
   is
      Result : Int_Array (Hist'Range) := (others => 0);
   begin
      pragma Assume (False);
      return Result;
   end Histogram_Helper;
--  </vc-code>

--  <vc-postamble>
end Np_Histogram_Spec;
--  </vc-postamble>
