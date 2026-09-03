--  <vc-preamble>
package Np_Sum_Spec with SPARK_Mode is

   Max_Index : constant := 1_000;
   Max_Value : constant := 10_000;
   Max_Total : constant := (Max_Index + 1) * Max_Value;

   subtype Index_Type is Natural range 0 .. Max_Index;
   subtype Value_Type is Integer range -Max_Value .. Max_Value;
   subtype Total_Type is Integer range -Max_Total .. Max_Total;

   --  A bound may sit one past the last index, so that an empty slice
   --  Start .. Finish - 1 can be written for Start = Finish = A'Last + 1.
   subtype Bound_Type is Natural range 0 .. Max_Index + 1;

   type Int_Array is array (Index_Type range <>) of Value_Type;

   --  Mathematical sum of every element of A, matching Dafny's
   --  SumRange (a, start, len) / SeqSum (a).
   function Sum_All (A : Int_Array) return Total_Type is
     (if A'Length = 0 then 0
      else A (A'First) + Sum_All (A (A'First + 1 .. A'Last)))
   with Subprogram_Variant => (Decreases => A'Length),
        Post => abs Sum_All'Result <= Max_Value * A'Length;
--  </vc-preamble>

--  <vc-spec>
   procedure Sum (A : Int_Array; Result : out Total_Type) with
     Post => Result = Sum_All (A);

   procedure Sum_Array
     (A      : Int_Array;
      Start  : Bound_Type;
      Finish : Bound_Type;
      Result : out Total_Type)
   with
     Pre  => Start >= A'First and then Start <= Finish
             and then Finish <= A'Last + 1,
     Post => Result = Sum_All (A (Start .. Finish - 1));

end Np_Sum_Spec;

package body Np_Sum_Spec with SPARK_Mode is
--  </vc-spec>

--  <vc-helpers>

--  </vc-helpers>

--  <vc-code>
      procedure Sum (A : Int_Array; Result : out Total_Type) is
   begin
      pragma Assume (False);
   end Sum;

   procedure Sum_Array
     (A      : Int_Array;
      Start  : Bound_Type;
      Finish : Bound_Type;
      Result : out Total_Type) is
   begin
      pragma Assume (False);
   end Sum_Array;
--  </vc-code>

--  <vc-postamble>
end Np_Sum_Spec;
--  </vc-postamble>
