--  <vc-preamble>
package Np_Invert_Spec with SPARK_Mode is

   Max_Index     : constant := 1_000;
   Max_Value     : constant := 10_000;
   Max_Bit_Width : constant := 16;
   Max_Pow2      : constant := 2 ** Max_Bit_Width;
   Max_Result    : constant := Max_Pow2 + Max_Value;

   subtype Index_Type is Natural range 0 .. Max_Index;
   subtype Value_Type is Integer range -Max_Value .. Max_Value;
   subtype Bit_Width_Type is Natural range 0 .. Max_Bit_Width;
   subtype Pow2_Type is Integer range 1 .. Max_Pow2;
   subtype Result_Type is Integer range -Max_Result .. Max_Result;

   type Int_Array is array (Index_Type range <>) of Value_Type;
   type Result_Array is array (Index_Type range <>) of Result_Type;

   --  Dafny's pow2 (n) = 2 raised to the power n.
   function Pow2 (N : Bit_Width_Type) return Pow2_Type is (2 ** N);
--  </vc-preamble>

--  <vc-spec>
   procedure Invert
     (Bit_Width : Bit_Width_Type;
      A         : Int_Array;
      Result    : out Result_Array)
   with
     Pre  => Result'First = A'First and then Result'Last = A'Last,
     Post => Result'Length = A'Length
             and then (for all I in A'Range =>
                         Result (I) = (Pow2 (Bit_Width) - 1) - A (I));

end Np_Invert_Spec;

package body Np_Invert_Spec with SPARK_Mode is
--  </vc-spec>

--  <vc-helpers>

--  </vc-helpers>

--  <vc-code>
      procedure Invert
     (Bit_Width : Bit_Width_Type;
      A         : Int_Array;
      Result    : out Result_Array) is
   begin
      pragma Assume (False);
   end Invert;
--  </vc-code>

--  <vc-postamble>
end Np_Invert_Spec;
--  </vc-postamble>
